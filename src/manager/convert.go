package manager

// 将joystick库输出的合并后的button值拆分回正常的多个button id的状态
func ConvertButton(button uint32) (buttons []int) {
	var result []int
	var m = button
	var count = 0
	for m > 0 {
		if m&1 == 1 {
			result = append(result, count)
		}
		m >>= 1
		count += 1
	}
	return result
}

// 将joystick库输出的-32767 - 32768范围的轴转为0 - 255范围的switch用的轴
func ConvertAxis(value int) int {
	result := (value + 32767) / 256

	// 消除摇杆抖动
	difference := result - 128
	if result > 128 && difference <= 5 {
		result = 128
	} else if result < 128 && difference >= -5 {
		result = 128
	}

	return result
}
