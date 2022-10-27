package http

import "testing"

func Test_getHTTPMethod(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name       string
		args       args
		wantMethod string
		wathPath   string
	}{
		{
			name: "get method",
			args: args{
				data: []byte(`GET /interactions/wait HTTP/1.1
			Host: localhost:54783
			User-Agent: Go-http-client/1.1
			Accept-Encoding: gzip`),
			},
			wantMethod: "GET",
			wathPath:   "/interactions/wait",
		},
		{
			name: "post method",
			args: args{
				data: []byte(`POST /v1/reports/9845a311-94a2-4bf0-bc3d-864bf2186b65?version=0 HTTP/1.1
			Host: localhost:54783
			User-Agent: Go-http-client/1.1
			Accept-Encoding: gzip`),
			},
			wantMethod: "POST",
			wathPath:   "/v1/reports/9845a311-94a2-4bf0-bc3d-864bf2186b65?version=0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMethod, gotPath := getHTTPMethodAndPath(tt.args.data)
			if gotMethod != tt.wantMethod {
				t.Errorf("gotMethod = %v, want %v", gotMethod, tt.wantMethod)
			}
			if gotPath != tt.wathPath {
				t.Errorf("gotPath = %v, want %v", gotPath, tt.wathPath)
			}
		})
	}
}

func TestNoOptionsSuccess(t *testing.T) {
	given, _, then := httpTest(t)
	given.
		a_http_toxic(nil).and().
		a_http_server()

	then.
		a_http_call_succeeds("/test", "GET")
}

func TestFailureOn(t *testing.T) {
	// option := map[string]Condition{
	// 	"/test": {
	// 		FailOn:       2,
	// 		RecoverAfter: 2,
	// 		Method:       "GET",
	// 	},
	// }

	given, when, then := httpTest(t)
	given.
		a_http_toxic(nil).and().
		a_http_server()

	when.
		a_http_call_succeeds("/test", "GET")

	then.
		a_http_call_succeeds("/test", "GET")
	//a_http_call_succeeds("/test", "GET")
}
