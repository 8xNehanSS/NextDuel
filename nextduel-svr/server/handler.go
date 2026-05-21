package server

type Handler func(ctx *Context) error

// func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	ctx := &Context{
// 		writer:  w,
// 		request: r,
// 	}
// 	if err := h(ctx); err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 	}
// }
