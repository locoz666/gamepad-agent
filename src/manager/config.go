package manager

import (
	"bufio"
	"fmt"
	"github.com/0xcafed00d/joystick"
	"github.com/deckarep/golang-set"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"log"
	"os"
	"reflect"
	"time"
)

type ActionInfo struct {
	Type        string // 按键button / 轴 axis
	Index       int    // 在对应的操作列表中处于第几位
	NormalValue int    // 给方向键用的，在使用单个轴区分上下或左右的情况下需要判断最大和最小值分别对应哪个方向
}

type ButtonMap struct {
	L      ActionInfo
	R      ActionInfo
	LS     ActionInfo // 指两个摇杆按下去的那个键
	RS     ActionInfo // 指两个摇杆按下去的那个键
	HOME   ActionInfo // 指PS键、XBOX键之类的按键
	SELECT ActionInfo
	START  ActionInfo
	B      ActionInfo
	A      ActionInfo
	Y      ActionInfo
	X      ActionInfo
}

type JoyStickMap struct {
	LsX ActionInfo // 左摇杆的X轴
	LsY ActionInfo // 左摇杆的Y轴
	RsX ActionInfo // 右摇杆的X轴
	RsY ActionInfo // 右摇杆的Y轴

}

type TriggerMap struct {
	ZL ActionInfo // 指L和R后面的那个扳机键
	ZR ActionInfo // 指L和R后面的那个扳机键
}

type ArrowMap struct {
	UP    ActionInfo
	DOWN  ActionInfo
	LEFT  ActionInfo
	RIGHT ActionInfo
}

type KeyCombination struct {
	InKeys []string
	OutKey string
}

type ServerConfig struct {
	Listen int // 监听的端口
}

type ClientConfig struct {
	ServerHost      string // 服务端的IP
	ServerPort      int    // 服务端的端口
	ButtonMap       ButtonMap
	JoyStickMap     JoyStickMap
	TriggerMap      TriggerMap
	ArrowMap        ArrowMap
	KeyCombinations map[string]KeyCombination
}

type Configuration struct {
	Type   string // server / client
	Server ServerConfig
	Client ClientConfig
}

var Config Configuration
var buttonMap, joyStickMap, triggerMap, arrowMap map[string]ActionInfo
var buttonKeys, joyStickKeys, triggerKeys, arrowKeys map[int]string
var js joystick.Joystick

func loadConfig() {
	var err error
	if err = viper.ReadInConfig(); err != nil {
		log.Fatalf("读取配置文件失败, %v", err)
	}
	if err = viper.Unmarshal(&Config); err != nil {
		log.Fatalf("配置文件解析失败, %v", err)
	}

	// client端做按键转换用，server端不需要
	if Config.Type == "client" {
		buttonMap, buttonKeys = MapConfig2Map("Client.ButtonMap", Config.Client.ButtonMap)
		joyStickMap, joyStickKeys = MapConfig2Map("Client.JoyStickMap", Config.Client.JoyStickMap)
		triggerMap, triggerKeys = MapConfig2Map("Client.TriggerMap", Config.Client.TriggerMap)
		arrowMap, arrowKeys = MapConfig2Map("Client.ArrowMap", Config.Client.ArrowMap)
	}
}

// 检测出扳机键是哪几个
func checkTrigger() (triggers []int) {
	tmp := mapset.NewSet()
	for i := 0; i < 10; i++ {
		state, err := js.Read()
		if err != nil {
			log.Panicf("读取手柄操作失败: %v", err)
		}
		for index, value := range state.AxisData {
			if value == -32767 {
				tmp.Add(index)
			}
		}

		time.Sleep(time.Millisecond)
	}
	for _, value := range tmp.ToSlice() {
		triggers = append(triggers, value.(int))
	}
	return triggers
}

// 创建扳机键的映射
func setTriggerMap(triggers []int) {
	t := reflect.TypeOf(Config.Client.TriggerMap)
	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i).Name
		log.Printf("请将 %s 扳机键按到底", key)

		isSuccess := false
		lastIndex := -1
		for {
			state, err := js.Read()
			if err != nil {
				log.Panicf("读取手柄操作失败: %v", err)
			}
			if isSuccess {
				if state.AxisData[lastIndex] == -32767 { // 确认扳机键是否被松开，如果被松开了就跳出去进行下一个按键的检测
					break
				} else {
					continue
				}
			}
			for _, axisIndex := range triggers {
				if state.AxisData[axisIndex] == 32768 {
					viper.Set(fmt.Sprintf("Client.TriggerMap.%s", key), ActionInfo{
						Type:  "axis",
						Index: axisIndex,
					})
					log.Printf("%s 扳机键已被设为 A%d", key, axisIndex)
					isSuccess = true
					lastIndex = axisIndex
				}
			}
			time.Sleep(time.Millisecond)
		}
	}
}

// 创建其他按键的映射
func setButtonMap() {
	t := reflect.TypeOf(Config.Client.ButtonMap)
	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i).Name

		log.Printf("请按下手柄上的 %s 键", key)

		isSuccess := false
		for {
			state, err := js.Read()
			if err != nil {
				log.Panicf("读取手柄操作失败: %v", err)
			}

			buttons := ConvertButton(state.Buttons)
			if isSuccess {
				if len(buttons) == 0 { // 确认按键是否被松开，如果被松开了就跳出去进行下一个按键的检测
					break
				} else {
					continue
				}
			}
			if len(buttons) == 1 {
				viper.Set(fmt.Sprintf("Client.ButtonMap.%s", key), ActionInfo{
					Type:  "button",
					Index: buttons[0],
				})
				log.Printf("%s 键已被设为 B%d", key, buttons[0])
				isSuccess = true
			} else if len(buttons) > 1 {
				log.Println("请勿同时按下多个键！")
			} else {
				continue
			}
			time.Sleep(time.Millisecond)
		}
	}
}

// 创建方向键的映射
func setArrowMap(triggers []int) {
	filterAxes := []int{0, 1, 2, 3} // 过滤掉摇杆
	filterAxes = append(filterAxes, triggers...)

	t := reflect.TypeOf(Config.Client.ArrowMap)
	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i).Name

		log.Printf("请按下手柄上的 %s 键", key)

		isSuccess := false
		var actionType string
		for {
			state, err := js.Read()
			if err != nil {
				log.Panicf("读取手柄操作失败: %v", err)
			}

			// 先检查是否有按键
			buttons := ConvertButton(state.Buttons)
			if isSuccess && actionType == "button" {
				if len(buttons) == 0 { // 确认按键是否被松开，如果被松开了就跳出去进行下一个按键的检测
					break
				} else {
					continue
				}
			}
			if len(buttons) == 1 {
				actionType = "button"
				viper.Set(fmt.Sprintf("Client.ArrowMap.%s", key), ActionInfo{
					Type:  actionType,
					Index: buttons[0],
				})
				log.Printf("%s 键已被设为 B%d", key, buttons[0])
				isSuccess = true
			} else if len(buttons) > 1 {
				log.Println("请勿同时按下多个键！")
			}

			// 如果按键没有被检测到的话，就看看轴有没有被打满的，有的话就是把方向键变为轴的手柄
			var minValue, minValueIndex, maxValue, maxValueIndex, axisIndex, normalValue int
			for index, value := range state.AxisData {
				// 判断是否是扳机键和摇杆，是的话就跳过
				if InSlice(index, filterAxes) {
					continue
				}
				// 判断是否是最小的元素
				if value < minValue {
					minValue = value
					minValueIndex = index
				} else if value > maxValue { // 判断是否是最大的元素
					maxValue = value
					maxValueIndex = index
				} else if (value == minValue && minValue != 0) || (value == maxValue && maxValue != 0) { // 判断是否出现了两个最小或最大值的轴，是的话可能是多个按键被同时按下了
					log.Println("请勿同时按下多个键！")
				}
			}
			if isSuccess && actionType == "axis" {
				if minValue == 0 && maxValue == 0 { // 确认按键是否被松开，如果被松开了就跳出去进行下一个按键的检测
					break
				} else {
					continue
				}
			}
			// 判断是否最大和最小值的轴都出现了，是的话可能是多个按键被同时按下了
			if minValue == -32767 && maxValue == 32768 {
				log.Println("请勿同时按下多个键！")
			} else if minValue == -32767 {
				axisIndex = minValueIndex
				normalValue = minValue
			} else if maxValue == 32768 {
				axisIndex = maxValueIndex
				normalValue = maxValue
			}
			if axisIndex != 0 {
				actionType = "axis"
				// 如果没有任何button操作并且还isSuccess了的话，就是把方向键变为轴的手柄上的方向键被按了
				viper.Set(fmt.Sprintf("Client.ArrowMap.%s", key), ActionInfo{
					Type:        actionType,
					Index:       axisIndex,
					NormalValue: normalValue,
				})
				log.Printf("%s 键已被设为 A%d", key, axisIndex)
				isSuccess = true
			}

			time.Sleep(time.Millisecond)
		}
	}
}

// 引导用户进行按键映射配置
func initJoystickMap() {
	// 由于viper暂时无法做到 直接设置默认配置 -> set一下特定值 -> 再合并回去 这样的操作，只能这么实现
	viper.Set("Client.JoyStickMap.LsX", ActionInfo{
		Type:  "axis",
		Index: 0,
	})
	viper.Set("Client.JoyStickMap.LsY", ActionInfo{
		Type:  "axis",
		Index: 1,
	})
	viper.Set("Client.JoyStickMap.RsX", ActionInfo{
		Type:  "axis",
		Index: 2,
	})
	viper.Set("Client.JoyStickMap.RsY", ActionInfo{
		Type:  "axis",
		Index: 3,
	})

	js = GetJoystickObject()

	triggers := checkTrigger()
	if len(triggers) > 0 {
		log.Println("你的手柄似乎存在带按键行程的扳机键，请输入y或n进行确认：")
		input := bufio.NewScanner(os.Stdin)
		for input.Scan() {
			inputText := input.Text()
			if inputText == "y" {
				setTriggerMap(triggers)
				break
			} else if inputText == "n" {
				break
			} else {
				log.Println("输入有误，请重新输入")
			}
		}
	}

	setButtonMap()
	setArrowMap(triggers)
}

// 初始化配置
func initConfig() {
	log.Println("未找到配置文件，开始初始化配置")

	// TODO 后面搞成一个DEBUG模式？有些手柄不按套路出牌
	// js = GetJoystickObject()
	// for {
	// 	state, err := js.Read()
	// 	if err != nil {
	// 		log.Panicf("读取手柄操作失败: %v", err)
	// 	}
	// 	axes := state.AxisData
	// 	buttons := ConvertButton(state.Buttons)
	// 	fmt.Println(buttons, axes)
	//
	// 	time.Sleep(time.Millisecond * 100)
	// }

	// 由于viper暂时无法做到 直接设置默认配置 -> set一下特定值 -> 再合并回去 这样的操作，只能这么设置默认配置
	viper.Set("Type", "client")
	viper.Set("Server.Listen", 9999)
	viper.Set("Client.ServerHost", "127.0.0.1")
	viper.Set("Client.ServerPort", 9999)

	initJoystickMap()

	err := viper.SafeWriteConfig()
	if err != nil && err != err.(viper.ConfigFileNotFoundError) {
		log.Fatalf("初始化失败, %v", err)
	}
}

// 导入包时会自动执行，解析配置文件并写进Config变量里
func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // 搜索配置文件的路径，可添加多个

	err := viper.ReadInConfig()
	if err != nil && err == err.(viper.ConfigFileNotFoundError) {
		initConfig()
	}

	loadConfig()
	log.Println("配置已加载")

	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		loadConfig()
		log.Printf("配置文件已重载")
	})
}
