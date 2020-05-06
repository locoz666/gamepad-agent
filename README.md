# gamepad-agent
一个纯 Go 实现的游戏手柄网络转发工具（实际上用的库并不纯）。

目前仅支持将 PC 端任意手柄转发给 Nintendo Switch，Switch 端依赖于 [mumumusuc](https://github.com/mumumusuc) 的
 [pi-joystick](https://github.com/mumumusuc/pi-joystick) 项目。

## 硬件方面
你需要准备：
- 1 台放在底座上或是外接了 USB HUB 的 Nintendo Switch
- 1 个树莓派 Zero W，用于转发手柄操作给 Switch
- 1 台连接好手柄的 PC，用于接收手柄操作并转发给树莓派
- 2 根 USB-A 转 Micro-USB 的线，一根用于给树莓派供电，一根用于将树莓派连接到 Switch 上
- 1 张 Micro SD 卡，用于给树莓派作为系统盘
- 1 个 5V 1A 或 5V 2A 的电源适配器，用于给树莓派供电（通常手机充电器就符合要求）

注意：树莓派 Zero 的普通版不带 WIFI 和蓝牙，只有 W 版和 WH 版才带。WH 和 W 的区别在于多了个 GPIO 排针，本项目中不需要用到 GPIO。

注意：PC 连接手柄时可以使用蓝牙连接也可以使用 USB 连接，但更推荐你使用 USB 连接。毕竟非内网环境使用的话，转发本身存在一定延迟，会导致体验下降。

注意：转发设备并不是一定要用树莓派 Zero W，其他支持 OTG 功能的开发板理论上也是可以的，只是**有可能**会需要自行编译一下 pi-joystick 和 gamepad-agent。

## 使用方法
客户端（PC）：
1. 下载或自行编译 客户端设备对应系统和架构版本的 gamepad-agent
2. 解压下载到的压缩包，进入解压出来的 gamepad-agent 目录下
3. 按照 example_config 中的样例配好配置文件，将配置文件中的 type 值设为 client
4. 在连接好手柄的状态下运行 gamepad-agent 启动客户端

服务端（树莓派）：
1. 下载或自行编译 服务端设备对应系统和架构版本的 gamepad-agent
2. 在树莓派上解压好下载到的压缩包，进入解压出来的 gamepad-agent 目录下
3. 执行 `sudo insmod js-audio.ko`，装载 pi-joystick
4. 把你的配置文件放进这个目录下，将配置文件中的 type 值设为 server
5. 执行 `chmod +x gamepad-agent` 给服务端文件添加执行权限
6. 执行 `./gamepad-agent` 启动服务端

双端启动好了之后，不出意外的话，你在 PC 上连接的手柄进行操作时，Switch 上就会收到对应的映射后的操作了。

## 样例配置
目前已经有了 ds4 手柄和 xbox360 手柄的映射配置，可以直接使用。在使用 moonlight 进行云游戏时，无论 moonlight 客户端设备使用的时什么手柄，都
可以直接使用 xbox360 手柄的映射配置。

## 原理
读取本机手柄状态 -> 转换为 Switch 的通信协议 -> 通过 UDP 发送到树莓派 Zero W -> 树莓派模拟成 USB HID 设备并将收到的 data 发给 Switch -> 
Switch 将树莓派当成 Pro 手柄，正常进行操作

## TODO
- 服务端支持Switch蓝牙连接
- 完全的 USB 转发或蓝牙转发
