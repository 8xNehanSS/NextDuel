package login

import "nex-server/server"

func HandleLogout(ctx *server.Context) error {
	ctx.Logout()
	return ctx.Success(nil)
}
