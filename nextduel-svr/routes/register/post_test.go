package register

import "testing"

func TestValidate(t *testing.T) {
	cases := []struct {
		name string
		req  RegisterRequest
		ok   bool
	}{
		{"valid", RegisterRequest{"player_1", "longenough", "Player@Example.com "}, true},
		{"short username", RegisterRequest{"ab", "longenough", "a@b.co"}, false},
		{"bad characters", RegisterRequest{"bad name!", "longenough", "a@b.co"}, false},
		{"bad email", RegisterRequest{"player", "longenough", "not-an-email"}, false},
		{"display-name email", RegisterRequest{"player", "longenough", "Bob <bob@b.co>"}, false},
		{"short password", RegisterRequest{"player", "short", "a@b.co"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := tc.req
			msg := req.Validate()
			if (msg == "") != tc.ok {
				t.Fatalf("Validate() = %q, want ok=%v", msg, tc.ok)
			}
		})
	}

	req := RegisterRequest{" player_1 ", "longenough", " Player@Example.COM"}
	req.Validate()
	if req.Username != "player_1" || req.Email != "player@example.com" {
		t.Fatalf("not normalised: %+v", req)
	}
}
