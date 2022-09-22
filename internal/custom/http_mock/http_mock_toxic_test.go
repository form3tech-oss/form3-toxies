package http_mock

import "testing"

func TestProxy(t *testing.T) {
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

func TestFailOnPathAndAttempt(t *testing.T) {
	given, when, then := HTTPMockTest(t)

	given.
		an_server_accepting_any_requests().and().
		an_http_mock_toxic_with_options(options{Path: "GET /accounts/id", FailOn: 2, RecoverAfter: 2}).and().
		an_get_request_with_path("/accounts/id")

	when.
		request_is_sent_n_times(4)

	then.
		request_failed_on_attempt(2)
}

func TestSucceedOnWrongPath(t *testing.T) {
	given, when, then := HTTPMockTest(t)

	given.
		an_server_accepting_any_requests().and().
		an_http_mock_toxic_with_options(options{Path: "GET /accounts/id", FailOn: 2, RecoverAfter: 2}).and().
		an_get_request_with_path("/users/id")

	when.
		request_is_sent_n_times(4)

	then.
		requests_were_successful()
}
