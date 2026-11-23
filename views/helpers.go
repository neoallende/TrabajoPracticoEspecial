package views

import (
	"strconv"

	"github.com/a-h/templ"
	db "github.com/neouuuu/TrabajoPracticoEspecial/db/sqlc"
)

func PanelFormEvento(usuarioID string, esAdmin bool, evento *db.Evento) templ.Component {
	return PanelFormEventoView(usuarioID, esAdmin, evento, true)
}

func PanelFormEventoPage(usuarioID string, esAdmin bool, evento *db.Evento) templ.Component {
	return PanelFormEventoView(usuarioID, esAdmin, evento, false)
}

func PanelTickets(usuarioID string, tickets []db.ListarTicketsPorUsuarioRow) templ.Component {
	return PanelTicketsView(usuarioID, tickets, true)
}

func LayoutHTMX(titulo string, contenido templ.Component) templ.Component {
	return Layout(titulo, contenido)
}

func tituloEvento(evento *db.Evento) string {
	if evento != nil {
		return "Editar evento"
	}
	return "Nuevo evento"
}

func botonEvento(evento *db.Evento) string {
	if evento != nil {
		return "Guardar cambios"
	}
	return "Crear evento"
}

func idEvento(evento *db.Evento) string {
	if evento != nil {
		return strconv.Itoa(int(evento.ID))
	}
	return ""
}

func nombreEvento(evento *db.Evento) string {
	if evento != nil {
		return evento.Nombre
	}
	return ""
}

func capacidadEvento(evento *db.Evento) string {
	if evento != nil {
		return strconv.Itoa(int(evento.Capacidad))
	}
	return ""
}

func fechaEvento(evento *db.Evento) string {
	if evento != nil {
		return evento.Fecha.Format("2006-01-02T15:04")
	}
	return ""
}

func claseNotificacion(errorTipo bool) string {
	if errorTipo {
		return "estado error"
	}
	return "estado ok"
}

func porcentaje(reservas int64, capacidad int32) float64 {
	if capacidad <= 0 {
		return 0
	}
	return float64(reservas) / float64(capacidad) * 100
}
