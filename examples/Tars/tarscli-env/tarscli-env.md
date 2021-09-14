###### tarscli-env是为了便于将tars运行在superedge上，对[TARSK8S](https://github.com/TarsCloud/K8STARS/)中的tarscli做了一定的改动

###### docker image： whispers1204/tarscli-env:v1

###### 具体改动如下：

1.将start.sh替换为本目录下的start.sh

2.将[tarsproxy/tarsregistry](https://github.com/TarsCloud/K8STARS/blob/master/tarsproxy/tarsregistry.go)的GetRegistryClient函数中的：

```go
if err := StringToProxy(locator, "tars.tarsregistry.Registry", client); err != nil {
	return nil
}
```

修改为：

```go
tars_registry := os.Getenv("TARS_REGISTRY")
if tars_registry == "" {
	tars_registry = "tars.tarsregistry.Registry"
}
if err := StringToProxy(locator, tars_registry, client); err != nil {
	return nil
}
```

3.将[tarsproxy/tarsnotify](https://github.com/TarsCloud/K8STARS/blob/master/tarsproxy/notify.go)的GetNotifyClient函数中的:

```go
if err := StringToProxy(locator, "tars.tarsnotify.NotifyObj", client); err != nil {
	return nil
}
```

修改为：

```go
tars_notify := os.Getenv("TARS_NOTIFY")
if tars_notify == "" {
	tars_notify = "tars.tarsnotify.NotifyObj"
}
if err := StringToProxy(locator, tars_notify, client); err != nil {
	return nil
}
```





