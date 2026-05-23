package login

import "nextduel-svr/server"

func HandleLogout(ctx *server.Context) {
	ctx.Logout()
	ctx.Redirect("/login", 302)
}
