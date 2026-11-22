package auth

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
	fallos         int
	bloqueadoHasta time.Time
	ultimoFallo    time.Time
}

func nuevoAuthLimitador() *authLimitador {
	l := &authLimitador{
		intentos: make(map[string]*intentoRegistro),
	}
	// Rutina de limpieza periódica para evitar fugas de memoria
	go l.limpiezaPeriodica(10 * time.Minute)
	return l
}

func (l *authLimitador) clave(ip, email string) string {
	return strings.ToLower(strings.TrimSpace(ip)) + ":" + strings.ToLower(strings.TrimSpace(email))
}

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

	// Si expiró el bloqueo, resetear
	if !reg.bloqueadoHasta.IsZero() && ahora.After(reg.bloqueadoHasta) {
		delete(l.intentos, k)
		return false, 0
	}

	// Si los fallos ocurrieron hace más de 15 minutos sin llegar a bloquearse, expirar
	if ahora.Sub(reg.ultimoFallo) > 15*time.Minute {
		delete(l.intentos, k)
		return false, 0
	}

	return false, 0
}

func (l *authLimitador) registrarFallo(ip, email string) (bloqueado bool, restante time.Duration) {
	l.Lock()
	defer l.Unlock()

	k := l.clave(ip, email)
	reg, existe := l.intentos[k]
	ahora := time.Now()

	if !existe {
		reg = &intentoRegistro{
			fallos:      0,
			ultimoFallo: ahora,
		}
		l.intentos[k] = reg
	}

	// Si el último fallo fue hace más de 15 min y no estaba bloqueado, reiniciar contador
	if reg.bloqueadoHasta.IsZero() && ahora.Sub(reg.ultimoFallo) > 15*time.Minute {
		reg.fallos = 0
	}

	reg.fallos++
	reg.ultimoFallo = ahora

	// 5 fallos => Bloqueo por 15 minutos
	if reg.fallos >= 5 {
		reg.bloqueadoHasta = ahora.Add(15 * time.Minute)
		return true, 15 * time.Minute
	}

	return false, 0
}

func (l *authLimitador) registrarExito(ip, email string) {
	l.Lock()
	defer l.Unlock()
	delete(l.intentos, l.clave(ip, email))
}

func (l *authLimitador) limpiezaPeriodica(intervalo time.Duration) {
	ticker := time.NewTicker(intervalo)
	for range ticker.C {
		l.Lock()
		ahora := time.Now()
		for k, reg := range l.intentos {
			if !reg.bloqueadoHasta.IsZero() && ahora.After(reg.bloqueadoHasta) {
				delete(l.intentos, k)
			} else if reg.bloqueadoHasta.IsZero() && ahora.Sub(reg.ultimoFallo) > 30*time.Minute {
				delete(l.intentos, k)
			}
		}
		l.Unlock()
	}
}
