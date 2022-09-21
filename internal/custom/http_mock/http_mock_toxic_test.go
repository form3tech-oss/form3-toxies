package http_mock

import (
	"fmt"
	"testing"
)

func TestProxy(t *testing.T) {
	t.Parallel()
	given, when, then := HTTPMockTest(t)

	given.
		an_server_accepting_any_requests().and().
		an_http_mock_toxic().and().
		an_get_request()

	when.
		request_is_sent()

	then.
		request_was_successful()
}

func TestStartFailingOnAttempt(t *testing.T) {
	given, when, then := HTTPMockTest(t)

	given.
		an_server_accepting_any_requests().and().
		an_http_mock_toxic_with_options(options{FailOn: 3}).and().
		an_get_request()

	when.
		request_is_sent_n_times(4)

	then.
		request_started_failing_on_attempt(3)
}

func TestFailOnAttemptAndRecover(t *testing.T) {
	t.Parallel()
	given, when, then := HTTPMockTest(t)

	given.
		an_server_accepting_any_requests().and().
		an_http_mock_toxic_with_options(options{FailOn: 2, RecoverAfter: 2}).and().
		an_get_request()

	when.
		request_is_sent_n_times(4)

	then.
		request_failed_on_attempt(2)
}

func TestSucceedOnWrongPath(t *testing.T) {
	t.Parallel()
	given, when, then := HTTPMockTest(t)

	given.
		an_server_accepting_any_requests().and().
		an_http_mock_toxic_with_options(options{Method: "GET", Path: "/accounts/id", FailOn: 2, RecoverAfter: 2}).and().
		an_get_request_with_path("/users/id")

	when.
		request_is_sent_n_times(4)

	then.
		requests_were_successful()
}

func TestFailOnAttempt(t *testing.T) {
	t.Parallel()
	cases := []struct {
		method  string
		path    string
		reqPath string
		failOn  int
	}{
		{
			method:  "GET",
			path:    "/accounts/.*?",
			reqPath: "/accounts/491dde0b-5a81-40c5-984a-ddb49032262a/status",
			failOn:  2,
		},
		{
			method:  "GET",
			path:    "/accounts/.*?/status",
			reqPath: "/accounts/491dde0b-5a81-40c5-984a-ddb49032262a/status",
			failOn:  1,
		},
		{
			method:  "POST",
			path:    "/accounts",
			reqPath: "/accounts",
			failOn:  3,
		},
		{
			method:  "PUT",
			path:    "/accounts/.*?",
			reqPath: "/accounts/b2691b66-a82f-4632-a40d-855d58b34d5e",
			failOn:  4,
		},
		{
			method:  "DELETE",
			path:    "/users/.*?",
			reqPath: "/users/59701156-4808-466e-b243-158b590e99d4",
			failOn:  3,
		},
	}
	for _, tt := range cases {
		t.Run(fmt.Sprintf("%s %s", tt.method, tt.path), func(t *testing.T) {
			given, when, then := HTTPMockTest(t)

			given.
				an_server_accepting_any_requests().and().
				an_http_mock_toxic_with_options(options{Path: tt.path, FailOn: tt.failOn, RecoverAfter: tt.failOn}).and().
				an_request(tt.method, tt.path)

			when.
				request_is_sent_n_times(4)

			then.
				request_failed_on_attempt(tt.failOn)
		})

	}

}
