package settings

type RemoteChromeSettings struct {
	Enable            bool   `json:"enable"`
	RemoteDockerURL   string `json:"remote_docker_url"`    // 整个是 go-rod 的远程镜像容器地址 ws://192.168.50.135:9222
	RemoteAdblockPath string `json:"remote_adblock_path"`  // 注意这个 go-rod 的远程镜像容器对应的目录， ADBlock 需要展开成文件夹 /mnt/share/adblock1_2_3
	ReMoteUserDataDir string `json:"remote_user_data_dir"` // 注意这个 go-rod 的远程镜像容器对应的目录，用户缓存文件地址
	// 注意，本程序没有办法去清理远程 docker 的文件夹的，这个请自行想办法
	// go-rod 的清理功能我测试是不正确的，作者的意思的是等系统重启清理系统的缓存，如果你常年挂机，应该就还是有问题的。
	// 其实你都映射出来了，完全可以定时用脚本在外部清理（但是可能本程序正在运行，所以···冲突，也就是为啥这个功能没有早期开放出来）
}
