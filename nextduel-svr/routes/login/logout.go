package login

import "nextduel-svr/server"

func HandleLogout(ctx *server.Context) error {
	ctx.Logout()
	ctx.Redirect("/login")
	return nil
}
