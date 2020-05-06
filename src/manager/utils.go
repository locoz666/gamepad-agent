package manager

import (
	"fmt"
	"github.com/0xcafed00d/joystick"
	"github.com/spf13/viper"
	"log"
	"reflect"
	"strings"
	"time"
)

func GetJoystickObject() (js joystick.Joystick) {
	var err error
	// 有些系统下第一个手柄是0，有些系统下第一个手柄是1，直接for一下简单粗暴
	for i := 0; i < 2; i++ {
		js, err = joystick.Open(i)
		if err != nil {
			log.Printf("尝试打开手柄 %d 失败: %v", i, err)
			continue
		}
		log.Printf("尝试打开手柄 %d 成功", i)
		break
	}
	return js
}

func ReadJoystick(js joystick.Joystick, chanel chan joystick.State) {
	for true {
		state, err := js.Read()
		if err != nil {
			log.Panicf("读取手柄操作失败: %v", err)
		}
		chanel <- state

		// 按照 https://github.com/dekuNukem/Nintendo_Switch_Reverse_Engineering 项目中的说法，15ms间隔比较合适
		// 原文 "only send out controller update every 15ms"
		time.Sleep(time.Millisecond * 15)
	}
}

func MapConfig2Map(configName string, configStruct interface{}) (mapResult map[string]ActionInfo, keysResult map[int]string) {
	mapResult = make(map[string]ActionInfo)
	keysResult = make(map[int]string)
	lowerKey2StructKey := map[string]string{}

	t := reflect.TypeOf(configStruct)
	v := reflect.ValueOf(configStruct)
	for i := 0; i < t.NumField(); i++ {
		key := t.Field(i).Name
		keysResult[int(v.Field(i).FieldByName("Index").Int())] = key
		lowerKey2StructKey[strings.ToLower(key)] = key
	}

	c := viper.Get(configName).(map[string]interface{})
	for actionName, _ := range c {
		var action ActionInfo
		err := viper.UnmarshalKey(fmt.Sprintf("%s.%s", configName, actionName), &action)
		if err != nil {
			log.Panicf("配置文件解析失败: %v", err)
		}
		mapResult[lowerKey2StructKey[actionName]] = action
	}
	return mapResult, keysResult
}

// 将joystick库输出的State转为标准化后的Action，方便后续操作
func JoystickState2Action(state joystick.State) (action Action) {
	mutable := reflect.ValueOf(&action).Elem()

	var axes []int
	for _, axe := range state.AxisData {
		axes = append(axes, ConvertAxis(axe))
	}
	buttons := ConvertButton(state.Buttons)

	// 其他按键
	for _, actionIndex := range buttons {
		if buttonName, ok := buttonKeys[actionIndex]; ok {
			action := buttonMap[buttonName]
			tmp := mutable.FieldByName(buttonName)
			switch action.Type {
			case "button":
				if InSlice(action.Index, buttons) {
					tmp.SetBool(true)
				}
			case "axis":
				axis := state.AxisData[action.Index]
				if action.NormalValue == axis {
					tmp.SetBool(true)
				}
			}
		}
	}
	// 摇杆
	for index, axisValue := range axes {
		if axisName, ok := joyStickKeys[index]; ok {
			mutable.FieldByName(axisName).SetInt(int64(axisValue))
		}
	}
	// 方向键
	for arrowName, action := range arrowMap {
		tmp := mutable.FieldByName(arrowName)
		switch action.Type {
		case "button":
			if InSlice(action.Index, buttons) {
				tmp.SetBool(true)
			}
		case "axis":
			axis := state.AxisData[action.Index]
			if action.NormalValue == axis {
				tmp.SetBool(true)
			}
		}
	}
	// 扳机键
	for triggerName, action := range triggerMap {
		tmp := mutable.FieldByName(triggerName)
		switch action.Type {
		case "button":
			if InSlice(action.Index, buttons) {
				tmp.SetBool(true)
			}
		case "axis":
			axis := state.AxisData[action.Index]
			if action.NormalValue == axis {
				tmp.SetBool(true)
			}
			// TODO 扳机键按压力度转发（目前Switch用不上，暂时不做）
		}
	}
	// 组合键处理
	for _, combination := range Config.Client.KeyCombinations {
		num := 0 // 有多少个组合内的按键被触发
		for _, key := range combination.InKeys {
			if mutable.FieldByName(strings.ToUpper(key)).Bool() {
				num += 1
			}
		}
		if num == len(combination.InKeys) {
			mutable.FieldByName(strings.ToUpper(combination.OutKey)).SetBool(true)
			// 完事之后把组合内的按键给设为false，以免误触
			for _, key := range combination.InKeys {
				mutable.FieldByName(strings.ToUpper(key)).SetBool(false)
			}
		}
	}
	return action
}

func InSlice(value int, list []int) bool {
	for _, tmp := range list {
		if tmp == value {
			return true
		}
	}
	return false
}

func Bool2Int(value bool) int {
	if value {
		return 1
	}
	return 0
}

// 将标准化后的Action转为switch手柄的通信协议
func Action2SwitchProtocol(action Action) []byte {
	buttonLow := Bool2Int(action.Y) | (Bool2Int(action.B) << 1) | (Bool2Int(action.A) << 2) | (Bool2Int(action.X) << 3) | (Bool2Int(action.L) << 4) | (Bool2Int(action.R) << 5) | (Bool2Int(action.ZL) << 6) | (Bool2Int(action.ZR) << 7)
	buttonHigh := Bool2Int(action.SELECT) | (Bool2Int(action.START) << 1) | (Bool2Int(action.LS) << 2) | (Bool2Int(action.RS) << 3) | (Bool2Int(action.HOME) << 4) | (Bool2Int(action.S1) << 5)
	hat := 8
	if action.UP {
		hat = 0
	}
	if action.RIGHT {
		hat = 2
	}
	if action.DOWN {
		hat = 4
	}
	if action.LEFT {
		hat = 6
	}
	result := []byte{uint8(buttonLow), uint8(buttonHigh), uint8(hat), uint8(action.LsX), uint8(action.LsY), uint8(action.RsX), uint8(action.RsY), 0}
	return result
}
