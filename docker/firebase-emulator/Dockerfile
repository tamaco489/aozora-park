FROM node:24.13.0-trixie-slim

# Firestore エミュレータの本体は Java で動く、firebase-tools が JDK 21 以上を要求するため trixie を使う
RUN apt-get update \
    && apt-get install -y --no-install-recommends default-jre-headless \
    && rm -rf /var/lib/apt/lists/*

RUN npm install -g firebase-tools@15.30.0

# 起動のたびに取りに行かせないため、エミュレータ本体と UI をイメージに焼き込む
RUN firebase setup:emulators:firestore \
    && firebase setup:emulators:ui

WORKDIR /workspace

# firebase.json は Hosting のデプロイと共有しているため、--only で hosting のエミュレータを巻き込まないようにする
CMD ["firebase", "emulators:start", "--only", "firestore", "--project", "demo-aozora-park"]
