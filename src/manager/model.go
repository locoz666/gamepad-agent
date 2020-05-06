package manager

// 标准化后的每次操作的状态
type Action struct {
	S1     bool // 预留四个特殊按键
	S2     bool // 预留四个特殊按键
	S3     bool // 预留四个特殊按键
	S4     bool // 预留四个特殊按键
	L      bool
	R      bool
	ZL     bool
	ZR     bool
	LS     bool
	RS     bool
	HOME   bool
	SELECT bool
	START  bool
	UP     bool
	DOWN   bool
	LEFT   bool
	RIGHT  bool
	B      bool
	A      bool
	Y      bool
	X      bool
	LsX    int
	LsY    int
	RsX    int
	RsY    int
}
