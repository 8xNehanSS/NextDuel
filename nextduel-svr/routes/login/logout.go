package login

import "nex-server/server"

func HandleLogout(ctx *server.Context) error {
	ctx.Logout()
	ctx.Redirect("/login")
	return nil
}
