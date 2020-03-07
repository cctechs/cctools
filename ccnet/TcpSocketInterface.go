package ccnet

type TcpSocketInterface interface {
	doWork()
	isAlive() bool
}