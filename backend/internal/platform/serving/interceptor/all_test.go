package interceptor

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"connectrpc.com/connect"

	parkv1 "github.com/tamaco489/aozora-park/backend/gen/aozorapark/park/v1"
	"github.com/tamaco489/aozora-park/backend/gen/aozorapark/park/v1/parkv1connect"
	"github.com/tamaco489/aozora-park/backend/internal/platform/serving/apperr"
)

// stubParkService は検証の対象がインターセプタの並びのため、業務処理を持たない
type stubParkService struct {
	parkv1connect.UnimplementedParkServiceHandler

	called bool
	err    error
}

var _ parkv1connect.ParkServiceHandler = (*stubParkService)(nil)

func (s *stubParkService) CreatePark(context.Context, *connect.Request[parkv1.CreateParkRequest]) (*connect.Response[parkv1.CreateParkResponse], error) {
	s.called = true
	if s.err != nil {
		return nil, s.err
	}
	return connect.NewResponse(&parkv1.CreateParkResponse{}), nil
}

func newTestClient(tb testing.TB, svc *stubParkService) parkv1connect.ParkServiceClient {
	tb.Helper()

	mux := http.NewServeMux()
	mux.Handle(parkv1connect.NewParkServiceHandler(svc, All(discardLogger())))

	server := httptest.NewServer(mux)
	tb.Cleanup(server.Close)

	return parkv1connect.NewParkServiceClient(server.Client(), server.URL)
}

func TestAll(t *testing.T) {
	valid := &parkv1.CreateParkRequest{
		Name:                 "Aozora Park",
		DefaultDailyCapacity: 100,
		InventoryDays:        14,
	}

	tests := map[string]struct {
		req        *parkv1.CreateParkRequest
		handlerErr error
		wantCode   connect.Code
		wantCalled bool
	}{
		"正常系_制約を満たす場合_ハンドラが呼ばれること": {
			req:        valid,
			wantCalled: true,
		},
		"異常系_nameが空の場合_ハンドラを呼ばずInvalidArgumentになること": {
			req:        &parkv1.CreateParkRequest{DefaultDailyCapacity: 100, InventoryDays: 14},
			wantCode:   connect.CodeInvalidArgument,
			wantCalled: false,
		},
		"境界値_inventory_daysが上限を超える場合_InvalidArgumentになること": {
			req:        &parkv1.CreateParkRequest{Name: "Aozora Park", DefaultDailyCapacity: 100, InventoryDays: 91},
			wantCode:   connect.CodeInvalidArgument,
			wantCalled: false,
		},
		"境界値_default_daily_capacityが0の場合_InvalidArgumentになること": {
			req:        &parkv1.CreateParkRequest{Name: "Aozora Park", DefaultDailyCapacity: 0, InventoryDays: 14},
			wantCode:   connect.CodeInvalidArgument,
			wantCalled: false,
		},
		"異常系_ハンドラがapperrを返す場合_Kindに対応するコードになること": {
			req:        valid,
			handlerErr: apperr.New(apperr.KindNotFound, "PARK_NOT_FOUND", "パークが見つからない"),
			wantCode:   connect.CodeNotFound,
			wantCalled: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			svc := &stubParkService{err: tt.handlerErr}
			client := newTestClient(t, svc)

			_, err := client.CreatePark(context.Background(), connect.NewRequest(tt.req))

			if tt.wantCode == 0 {
				if err != nil {
					t.Fatalf("CreatePark(%v) = %v, want nil", tt.req, err)
				}
			} else {
				connectErr, ok := errors.AsType[*connect.Error](err)
				if !ok {
					t.Fatalf("CreatePark(%v) = %v, want *connect.Error", tt.req, err)
				}
				if connectErr.Code() != tt.wantCode {
					t.Errorf("CreatePark(%v) のコード = %v, want %v", tt.req, connectErr.Code(), tt.wantCode)
				}
			}

			if svc.called != tt.wantCalled {
				t.Errorf("CreatePark(%v) のハンドラ呼び出し = %v, want %v", tt.req, svc.called, tt.wantCalled)
			}
		})
	}
}
