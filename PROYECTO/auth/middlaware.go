package auth

func EsAdmin(usuario models.usuario) bool {
	return usuario.Rol == "admin"
}
