package ccnet

type TcpSocketInterface interface {
	DoWork()
	IsAlive() bool
}