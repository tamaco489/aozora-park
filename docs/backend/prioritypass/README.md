# The Pub/Sub path of the priority pass

[English](./README.md) | [日本語](./README.ja.md)

Back to the [documentation index](../../README.md).

This page describes the Pub/Sub path that connects a request (`api`) to an allocation (`priority-pass-issuer`).
The rules live in `.claude/rules/go/coding.md` and `.claude/rules/infra/coding.md`; this page records the current shape and why it was chosen.

## What the feature does

A guest requests a priority pass for a time slot of an attraction. If the slot has capacity left the pass is issued, otherwise it ends as sold out.

| Status      | Meaning                                        | Written by             |
| ----------- | ---------------------------------------------- | ---------------------- |
| `requested` | The request was accepted and awaits allocation | `api`                  |
| `issued`    | A slot was taken and the pass was issued       | `priority-pass-issuer` |
| `sold_out`  | No capacity was left                           | `priority-pass-issuer` |

`api` never touches the inventory; it stores the pass as `requested`. Only the allocation takes a slot, so the request RPC does not depend on the state of the inventory.
A transition is allowed once, out of `requested`; both `issued` and `sold_out` are terminal.
A `rejected` status will be added once `tickets` exists and the pass can be checked against a ticket.

Every transition is appended to `events`. `actor` is `api` or `system:priority-pass-issuer`, `action` is `created` or `status_changed`,
and `cause` is `guest_requested`, `slot_allocated` or `slot_sold_out`. The status itself is carried by `changes`.

## The path

![The path from a request to an allocation](./images/flow.png)

| Hop                      | What happens                                                                                       |
| ------------------------ | -------------------------------------------------------------------------------------------------- |
| Guest → `api`            | `RequestPriorityPass` creates the pass and its `events` entry in one transaction                   |
| `api` → topic            | Publishes to `prioritypass.requested`, then fills in `publishedAt`                                 |
| Topic → subscription     | `prioritypass.requested.allocate` delivers it by push                                              |
| → `priority-pass-issuer` | `POST /pubsub/push`, authenticated as `sa-pubsub-push` with an OIDC token                          |
| → Firestore              | The transition, the decrement of `timeSlots` and the `events` entry are written in one transaction |
| Failed delivery → DLQ    | After five attempts the message moves to `prioritypass.requested.dlq`, held by `.dlq.hold`         |

The body carries nothing but `{"passId": "..."}`. Firestore is the source of truth for the request itself, and the subscriber reads it back by the identifier.
Adding fields would grow the contract even if the subscriber never reads them, so fields are added only when they are needed.

Only `sa-pubsub-push` can reach the worker: the ingress of `priority-pass-issuer` is internal only, and it is the one principal holding `roles/run.invoker` on it.
Cloud Run IAM and the OIDC token guarantee that the caller is Pub/Sub, so the handler does not verify it again.

The push path exists in two places: `local.push_path` in `infra/modules/pubsub/main.tf` and `pushPath` in `backend/cmd/priority-pass-issuer/main.go`. Changing one means changing the other.

## Why push and not pull

Pull needs a process that stays up to hold the subscription. `priority-pass-issuer` runs with `min_instance_count = 0`
so that no instance is paid for while there are no messages, which does not fit a model that assumes a resident subscriber.

With push the HTTP request from Pub/Sub is what starts an instance, which is exactly how Cloud Run scales.
The ack deadline is 60 seconds so that the work still finishes within it even after a cold start.

The endpoint is a plain HTTP handler rather than a connect RPC. Modelling it in connect would put the shape of the envelope into the proto and generate it for the frontend as well.
Push is also fixed to an HTTP/1.1 POST, so the gRPC and gRPC-Web paths of connect could not be used anyway.

## Ack and nack

Push has no nack API. **The status code of the response is the only thing that separates an ack from a redelivery.**
That decision is made by `internal/platform/serving/pubsubpush`; the feature handler only returns an `error`.

| What happened in the handler               | Response | Result                       |
| ------------------------------------------ | -------- | ---------------------------- |
| The work succeeded                         | `204`    | Ack                          |
| The envelope could not be decoded          | `204`    | Ack and log                  |
| An error where `apperr.Retryable` is false | `204`    | Ack and log                  |
| An error where `apperr.Retryable` is true  | `500`    | Redelivery                   |
| A method other than `POST`                 | `405`    | A caller that is not Pub/Sub |

Whether an error is retryable is decided where the sentinel is defined: `apperr.New` gives `false`, `apperr.NewRetryable` gives `true`.

| Error                                | Retryable | Why                                                     |
| ------------------------------------ | --------- | ------------------------------------------------------- |
| `PRIORITY_PASS_BAD_MESSAGE`          | `false`   | The same body will never decode, however often it comes |
| `PRIORITY_PASS_NOT_FOUND`            | `false`   | A pass that does not exist will not appear on a resend  |
| `PRIORITY_PASS_TIME_SLOT_NOT_FOUND`  | `true`    | The slot generation may simply not have caught up yet   |
| Anything unclassified (SDK, network) | `true`    | Redelivering is safer than swallowing it                |

Returning `500` for something that cannot be fixed repeats the same failure five times before it lands in the DLQ.
The attempts are wasted, and the DLQ ends up mixing what should be replayed with what should be dropped.

A missing slot is not folded into sold out, because that would stop the pass in a terminal state it cannot leave even once the slot exists.

## Why it is idempotent

Pub/Sub is at-least-once. **The handler assumes the same message arrives more than once and writes nothing the second time.**

The idempotency key is `passId`. Since the body carries nothing but the identifier, a redelivery always points at the same single pass.

The allocation runs as one Firestore transaction:

1. Read `priorityPasses/{passId}`
2. If the status is not `requested`, write nothing and return (`Changed` is `false`)
3. Read `remaining` of the `timeSlots` document
4. If `remaining > 0`, move to `issued` and decrement it by one; if it is `0`, move to `sold_out`
5. Write the update of `priorityPasses` and the `events` entry in the same transaction

Step 2 is what makes it idempotent. The domain model also only allows a transition out of `requested`, so there is no path back from a terminal status.
The second delivery returns successfully with `Changed` as `false`, and the endpoint acks it with `204`.

The decrement and the transition share a transaction because leaving only one of them would hand out the same slot twice.
This is the documented exception to "one aggregate per transaction", allowed for operations that touch the inventory.
`remaining` cannot go negative: a Firestore transaction retries when a document it read has changed, so subtracting from the value that was read is safe.

## Delivery settings

They live in `infra/modules/pubsub/main.tf`.

| Setting                          | Value                                         | Why                                                                               |
| -------------------------------- | --------------------------------------------- | --------------------------------------------------------------------------------- |
| `ack_deadline_seconds`           | `60`                                          | One transaction is enough work; leaves room for a cold start                      |
| `retry_policy`                   | `minimum_backoff` 10s, `maximum_backoff` 600s | Spaces out the retries of a failure that may still succeed                        |
| `max_delivery_attempts`          | `5`                                           | Keeps a failure that cannot be fixed from repeating forever                       |
| `expiration_policy.ttl`          | `""` (never)                                  | The default 31 days of inactivity would silently drop the target of every publish |
| DLQ `message_retention_duration` | `604800s` (7 days)                            | Enough time to notice and replay by hand; longer costs storage                    |

Only `prioritypass.requested.allocate` has a dead letter policy. `prioritypass.requested.dlq` does not get one of its own,
because nothing delivers from it and there would be nowhere further to move a message to.

Instead there is `prioritypass.requested.dlq.hold`, a pull subscription that exists only to retain messages.
A topic drops messages when it has no subscription, so without it nothing in the DLQ could be read afterwards.
It has no push endpoint because replaying is done by hand with `gcloud`.

Moving a message to the DLQ is done by the Pub/Sub service agent, which is granted `publisher` on the dead letter topic and `subscriber` on the source subscription.
That agent is not created by enabling the API, so it is created by hand once per environment.

## Detecting a publish that never happened

The publish is detached from the caller's request and capped at three seconds (`context.WithoutCancel` plus `context.WithTimeout`).
The pass already exists, so a failed publish does not roll it back; rolling it back and letting the caller retry would create a duplicate request.

A successful publish is recorded in `publishedAt` on the `priorityPasses` document. It starts as `null` and is filled in only after the publish result is confirmed.

| State on the document                               | Meaning                                 |
| --------------------------------------------------- | --------------------------------------- |
| `publishedAt` holds a timestamp                     | The publish was confirmed               |
| `publishedAt` is `null` and `status` is `requested` | Either the publish or its record failed |

> [!IMPORTANT]
> **The reconciliation that would resend them does not exist yet.** The only subcommand `backend/cmd/job` has is `generate`.
> What holds today is that the passes left with `publishedAt` as `null` can be listed; resending them is a manual step.

`publishedAt` records the publish rather than a domain state, so writing it leaves `status` and `updatedAt` untouched.
If only the record fails, the message was still sent, and a resend is handled idempotently by the subscriber anyway.

## Why there is no ordering key

`prioritypass.requested` uses no ordering key, and `enable_message_ordering` is not set.

- A message stands for one request, and requests have no order between them. A redelivery of the same `passId` is idempotent, so the order of processing does not change the outcome
- Who wins a contested slot is settled by the Firestore transaction, not by the order of messages, so there is nothing to ask of the delivery side
- With ordering enabled, messages sharing a key wait until the previous one is acked. One failure would then block everything behind it, which costs throughput and widens the blast radius

## Checking it locally

The emulators are defined in `docker-compose.yaml`.

```sh
just up    # Firestore (18080) and Pub/Sub (18085)
just down
```

A one-shot `pubsub-init` service creates `prioritypass.requested`, because the emulator does not create a topic on publish.

```sh
cd backend
just run-api  # starts the api against the emulators on 8080
```

Calling `RequestPriorityPass` publishes to the emulated topic and fills in `publishedAt`. That much can be checked with the emulators alone.

The receiving side is exercised by running `priority-pass-issuer` on another port and calling it directly.
No push subscription is created in the emulator, so nothing is delivered automatically.

```sh
cd backend
FIRESTORE_EMULATOR_HOST=localhost:18080 \
  GOOGLE_CLOUD_PROJECT=demo-aozora-park \
  PORT=8081 go run ./cmd/priority-pass-issuer
```

Build an envelope in the shape push uses, with `data` holding the base64 of the body.

```sh
DATA=$(printf '{"passId":"<the pass id>"}' | base64)
curl -i -X POST http://localhost:8081/pubsub/push \
  -H 'Content-Type: application/json' \
  -d "{\"message\":{\"messageId\":\"1\",\"data\":\"$DATA\"},\"subscription\":\"projects/demo-aozora-park/subscriptions/prioritypass.requested.allocate\"}"
```

A `204` on the first call and a `204` with no change in state on the second means it is idempotent. A `500` only appears for a failure worth redelivering.
The allocation needs `timeSlots`, so generate the inventory first with `go run ./cmd/job generate`.

The automated tests never start a Pub/Sub emulator. The shape of the body is checked against the function that builds it, and the endpoint is checked by passing an envelope through `httptest`.

## Checking it on stg

What cannot be reproduced locally — the internal-only ingress, the OIDC check, the move to the DLQ — is checked on stg.
The deployment procedure is in [Deploying the backend to stg](../../cd/backend/stg.md).

1. Merge into `main` and confirm that `cd-api-stg`, `cd-priority-pass-issuer-stg` and `cd-frontend-stg` all pass
2. Create a park from the UI and note its `parkId`
3. Create an attraction from the UI and note its `attractionId`. **Set `enabled` to true**
4. Generate the inventory with the command below
5. Request a pass from the UI, then check its state with the `passId` that comes back

```sh
cd backend
just generate-inventory-stg
```

> [!IMPORTANT]
> **Requesting a pass without generating the inventory fails in a way that is hard to read.** The request itself succeeds and returns `requested`,
> but the allocation fails with `PRIORITY_PASS_TIME_SLOT_NOT_FOUND`, is redelivered five times and ends up in the DLQ.
> The UI keeps showing `requested`, so nothing points at the cause until you look at the DLQ.

Slots are only generated for attractions whose `priorityPassConfig.enabled` is true, and only for `inventoryDays` days from today.
Creating a park or an attraction does not create any slots; `generate` in `cmd/job` is the only thing that does.

> [!NOTE]
> Nothing runs that generation on a schedule yet (no Cloud Run job, no Cloud Scheduler).
> For now the command above is run by hand. Putting it on CD is tracked in #135.

A `timeSlotId` is shaped `YYYYMMDD_HHMM`. The time slot screen shows the start time but not the identifier, so it is assembled from the date and the time.
A `ticketId` is never matched against a ticket, so any string will do.

## Naming

| Thing                      | Shape                             | Example                                     |
| -------------------------- | --------------------------------- | ------------------------------------------- |
| Topic                      | `<feature>.<event in past tense>` | `prioritypass.requested`                    |
| Subscription               | `<topic>.<what it does>`          | `prioritypass.requested.allocate`           |
| Dead letter topic          | `<topic>.dlq`                     | `prioritypass.requested.dlq`                |
| Subscription holding a DLQ | `<dead letter topic>.hold`        | `prioritypass.requested.dlq.hold`           |
| Cloud Run worker           | An `-er` noun                     | `priority-pass-issuer`                      |
| Service account            | `sa-<service>` or `sa-<role>`     | `sa-priority-pass-issuer`, `sa-pubsub-push` |

Event names are in the past tense because the meaning of an event that has already been published must not change later.
A subscription carries the topic in its name so that, once several subscriptions read one topic, the name alone says whose backlog is stuck.

GCP projects are split per environment, so no resource name carries an environment prefix.
