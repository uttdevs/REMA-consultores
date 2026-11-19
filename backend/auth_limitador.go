package main

import (
	"strings"
	"sync"
	"time"
)

// authLimitador administra el bloqueo por intentos fallidos de login:
// Regla: 5 fallidos → bloqueo temporal de 15 min por IP+email.
type authLimitador struct {
	sync.Mutex
	intentos map[string]*intentoRegistro
}

type intentoRegistro struct {
	fallos        int
	bloqueadoHasta time.Time
	ultimoFallo   time.Time
}

func nuevoAuthLimitador() *authLimitador {
	l := &authLimitador{
		intentos: make(map[string]*intentoRegistro),
	}
	// Limpieza periódica cada 10 minutos
	go l.limpiezaPeriodica(10 * time.Minute)
	return l
}

func (l *authLimitador) clave(ip, email string) string {
	return strings.ToLower(strings.TrimSpace(ip)) + "|" + strings.ToLower(strings.TrimSpace(email))
}

// estaBloqueado revisa si la combinación IP+email está bloqueada temporalmente.
func (l *authLimitador) estaBloqueado(ip, email string) (bool, time.Duration) {
	l.Lock()
	defer l.Unlock()

	k := l.clave(ip, email)
	reg, existe := l.intentos[k]
	if !existe {
		return false, 0
	}

	ahora := time.Now()
	if reg.bloqueadoHasta.After(ahora) {
		return true, reg.bloqueadoHasta.Sub(ahora)
	}

	// Si el bloqueo ya expiró, reseteamos el registro
	if !reg.bloqueadoHasta.IsZero() && ahora.After(reg.bloqueadoHasta) {
		delete(l.intentos, k)
	}

	return false, 0
}

// registrarFallo contabiliza un fallo. Si alcanza 5 fallos, activa el bloqueo de 15 min.
func (l *authLimitador) registrarFallo(ip, email string) (bloqueado bool, restante time.Duration) {
	l.Lock()
	defer l.Unlock()

	k := l.clave(ip, email)
	reg, existe := l.intentos[k]
	ahora := time.Now()

	if !existe {
		reg = &intentoRegistro{
			fallos:      1,
			ultimoFallo: ahora,
		}
		l.intentos[k] = reg
		return false, 0
	}

	// Si el último fallo fue hace más de 15 minutos y no estaba bloqueado, reiniciamos la cuenta
	if reg.bloqueadoHasta.IsZero() && ahora.Sub(reg.ultimoFallo) > 15*time.Minute {
		reg.fallos = 1
		reg.ultimoFallo = ahora
		return false, 0
	}

	reg.fallos++
	reg.ultimoFallo = ahora

	if reg.fallos >= 5 {
		reg.bloqueadoHasta = ahora.Add(15 * time.Minute)
		return true, 15 * time.Minute
	}

	return false, 0
}

// registrarExito borra el historial de fallos cuando el usuario se loguea correctamente.
func (l *authLimitador) registrarExito(ip, email string) {
	l.Lock()
	defer l.Unlock()
	delete(l.intentos, l.clave(ip, email))
}

func (l *authLimitador) limpiezaPeriodica(intervalo time.Duration) {
	ticker := time.NewTicker(intervalo)
	for ahora := range ticker.C {
		l.Lock()
		for k, reg := range l.intentos {
			// Si expiró el bloqueo o no ha habido actividad en más de 30 minutos
			if (!reg.bloqueadoHasta.IsZero() && ahora.After(reg.bloqueadoHasta)) ||
				(reg.bloqueadoHasta.IsZero() && ahora.Sub(reg.ultimoFallo) > 30*time.Minute) {
				delete(l.intentos, k)
			}
		}
		l.Unlock()
	}
}
