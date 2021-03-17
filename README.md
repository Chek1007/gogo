# Getitle
一个资产探测扫描器. 

## Usage

```
Usage of ./getitle:
  -d int       超时,默认2s (default 2)
  -ip string   IP地址 like 192.168.1.1/24
  -m string    扫描模式：default or s(smart)
  -n string    协议模式: socket(默认) or icmp
  -p string    ports (default "top1")
  -t int       threads (default 4000)
  -o string    输出格式:clean,full(default) or json
  -f string    输出文件名,默认为空,请用相对路径(./)或绝对路径
  -v bool      扫描详细指纹,当前主要有常见CMS指纹
  -e bool      启用漏洞插件扫描,目前主要有ms17-010与shiro
  -k string    启动密码(必须输入)为puaking  
  -l string    从文件中读取任务,例如:-l ip.txt
  -P bool      查看端口预设
  -F file      格式化json

```



### 用法

所有用法都需输入-k [密钥]

#### 扫描C段的关键端口

`./gt.exe -ip 192.168.1.1/24 -p top2`

#### 扫描启发式扫描B段或大于B段

`./gt.exe -ip 172.16.1.1/12 -p top2 -m s`

启发式扫描只会先扫描80端口,如果在该C段中扫描到了80端口,则进行已配置端口的完整扫描.加快扫描速度.

#### 使用端口预设

`./gt -P`

![image-20210317223831976](README.assets/image-20210317223831976.png)

根据类型名或者服务名自由组合端口(会自动去重),例如

`./gt.exe -ip 172.16.1.1/24 -p top2,db,mail,jboss,1000-1009,12345,54321`

#### 指纹识别

当前包括一百多条CMS指纹以及数十条favicon指纹.

用法:

`./gt.exe -ip 192.168.1.1/24 -p top2 -v`

#### 漏洞探测

getitle并非漏扫工具,因此不会支持sql注入,xss之类的漏洞探测功能.

当前漏洞只能针对简单的http或tcp连接交互实习.

当前支持漏洞:

shiro(默认key)

ms17-010

`./gt.exe -ip 192.168.1.1/24 -p top2 -e`

#### 高级启发式扫描

自动内网网段扫描,扫描全内网所有C段第一个ip,仅支持icmp

`./gt.exe -m a`



首ip网段探测,可通过-n 指定,默认为socket,例如:

`gt.exe -ip 59.1.1.1/8 -m f`



指定为首ip探测模式 icmp协议

`gt.exe -ip 59.1.1.1/8 -m f -n icmp`



#### 特殊扫描端口

当前包括三个特殊扫描端口(能够和其他端口配置组合不冲突)

OXID:

`./gt.exe -ip 172.16.1.1/24 -p top2,135`

NBTScan

`./gt.exe -ip 172.16.1.1/24 -p top2,137`

ICMP

`./gt.exe -ip 172.16.1.1/24 -p top2,icmp`





### 注意事项

* 因为并发过高,性能限制主要来自路由器设备.因此**建议在阿里云,华为云等vps上使用**,如果扫描国外资产,建议在国外vps上使用.本地使用如果网络设备性能不佳会带来大量丢包.

* 如果使用中发现疯狂报错,大概率是io设备问题(例如多次扫描后io没有被正确释放,或者配合proxifier以及类似代理工具使用报错),可以通过重启电脑,或者虚拟机中使用,关闭代理工具解决.如果依旧无法解决请联系我们.

* 还需要注意,upx压缩后的版本虽然体积小,但是有可能被杀软杀,也有可能在部分机器上无法运行.

* 一般情况下无法在代理环境中使用,除非使用-t参数指定较低的速率(默认协程池为4000).

### 实验性功能:

-e参数开启简单的漏洞探测.

当前只支持redis未授权,shiro key发现,ms17-010(免杀原因临时移除)

`./gt.exe -ip 172.16.1.1/12 -p top2,445  -e`

-v参数开启详细版本号匹配

`./gt.exe -ip 172.16.1.1/12 -p top2  -v`

icmp扫描

指定探测内网网段: `./gt.exe -m a -p top2 ` (默认为`10/8,172.16/12,192.168/16`)

外网icmp探测: `./gt.exe -ip 47.95.116.1/16 -p icmp`

外网icmp启发式扫描:  `./gt.exe -ip 47.95.116.1/16 -p top2 -m s -n icmp`



## 配置指纹识别

指纹的json位于`src\Utils\finger.json`.

为了保证单文件可使用,将会在运行gox.bat时将json中的数据写到`src\Utils\finger.go`中



配置示例:

```
[    
	{
        "name": "Mysql_unauthorized",
        "level": 0,
        "defaultport": "3306",
        "regexps": [
            "Host .* is not allowed to connect to this MySQL server"
        ]
    }
]
```

`name`为规则名,string,请保证不重名

`level`为优先级,int,最高优先级为0

`defaultport`为该服务默认端口,string,用作提高匹配速度

`regexps`为正则列表,[]string, 默认为数组,同一规则可以配置多个正则依次匹配

### 注意事项

* json不接受`\x00`,`\0`等转义,请将类似转义修改成`\u0000`.

* 请注意数组元素间的逗号,否则可能导致json报错

## Makefile

 * make release VERSION=VERSION to bulid getitle to all platform

 * Windows build muli releases

   ```
   go get github.com/mitchellh/gox
   gox.bat
   ```

   

## Change Note

* v0.0.1 just a demo

* v0.0.3 
  
  * 获取不到getitile的情况下输出前13位字符(如果是http恰好到状态码)
  
* v0.0.4 
  * 添加了端口预设top1为最常见的http端口,top2为常见的http端口,db为常见数据库默认端口,win为windows常见开放的端口
  * 简化了端口参数
  
* v0.0.5 
  * 修复了400与30x页面无法获取titile的问题
  * 修复了无法自定义端口的bug
  * 添加了brute与all两个端口预设,brute为可爆破端口,all为常见端口
  * 忽略匹配title的大小写问题
  
* v0.0.6
  
  * 添加了大于B段启发式扫描模式
  
* v0.1.0
  * 优化了参数
  * 添加了ms17010漏洞扫描
  * 修复了扫描单个ip报错的情况
  
* v0.1.1

  * 修复了启发式扫描的ip计算错误的bug
  * 添加了基于`Server`与`X-Powered-By`的简单指纹识别  
  
* v0.1.2
  * 添加了redis未授权扫描
  * 重构了输出函数
  
* v0.1.3
  * 添加了nbtscan
  * 修复了部分bug
  * 添加了json格式输出
  
* v0.1.4
  * 修复了rediscan未位置timeout的bug
  * 添加了更复杂的输出模式
  * 去除了banner
  * 添加了key
  
* v0.1.5
  * 添加了-f参数,输出到文件,会有30%的性能损失
  * 修复了格式化输出的一个bug,导致无法使用tee重定向
  
* v0.1.6

    * 修复了输出文件降低效率的问题(优化写入方式)
    * 优化了tee保存结果,去除进度信息保存到文件
    * 添加了OXID扫描
    * 优化了二进制文件大小
    * 添加了更强的端口配置模式,例如` -p top2,db,12345` 
    * 修复了无法扫描A段的bug
    
* v0.2.0(beta1/2)
    * 修复了OXID部分情况下报错的bug
    * 修复了https无法获取title的bug
    * 优化了匹配title,中间件与语言的正则表达式,使其可以匹配到更多的信息
    * 优化了端口配置,all预设将会拼接所有的预设并自动去重
    * 优化了输出格式
    * 优化了OXIDscan的网卡信息输出
    * 添加了shiroscan(beta)
    * 添加了-e参数(bool值,不需要添加值,只需要存在-e即可),redisscan与shiroscan只有在开启该参数才会启用(beta)
    
* v0.2.0(beta3)
    * 修复了https协议shiro无法正确识别的bug
    * 优化了Nbtscan于OXIDscan中不可见字符输出问题
    * 添加了top3端口预设
    * 使用go1.11编译,兼容windows server2003与xp系统
    
* v0.2.0(beta4)
    * 添加了证书信息收集
    * 添加了线程的安全的slice进行启发式扫描
    * 优化了扫描逻辑与扫描性能
    * 优化了扫描进度信息输出
    * 优化了内存占用,扫描A段内存占用低于150M
    * 修复了多个bug
    * 临时删除了ms17-010扫描
    
* v0.2.0(beta5)
  * 修复了32位版本ip解析错误的bug
  * 优化了top2,top3端口预设
  * 添加了-l参数,可从文件中读取任务
  * 优化了证书信息收集
  * 优化了http扫描,增加了https与跳转的请求超时时间.
  * 优化了文件写入,防止程序中途崩溃导致进度消失.
  * 修复了一个json格式输出的bug
  
* v0.2.0(beta6)
    * 现在ip参数可以接受`https://1.1.1.1`,并自动删除`https://`或`http://`
    * 现在ip参数可以接受域名如`https://baidu.com`,并自动获取ip.如果域名绑定了多个ip,只会获取第一个.
    * 优化了top2默认端口,添加了`1080,3000,5000,6443`等常见端口
    * -o 参数新增html格式,使用方式为`-o html`
    * 新增tcp端口指纹探测,-v参数启用,详细配置见`配置指纹识别`
    * 优化了输出信息,更准确的判断http与tcp协议.
    * 修复子网掩码为8/9的情况下的ip解析错误
    
* v0.2.0(beta7)
    * 新增-c(clean)参数,如果存在则命令行只输出进度信息
    * 重新添加-k,默认key为`puaking`
    * 修复特定情况下证书错误的bug
    * 新增-ip参数的`auto`关键字,将会自动对`10/8,172.16/12,192.168/16`三个内网网段的所有c段的第一个ip,如10.0.0.1进行探测,示例`./gt -ip auto -p top2`,默认为icmp扫描,也可以使用-m s指定为socket探测80端口
    * 新增icmp扫描,有三种打开方式
        1. 在-p参数中添加icmp,例如`-p top2,icmp`
        2. 在-m参数指定`-m sp`,则使用icmp进行启发式扫描
        3. 在-ip为`auto`的时候,自动启用icmp探测

* v0.2.0(beta7.1)
    修复beta7版本的多个bug

    

​    



 ## Todo List

1. 添加NetBIOS  [√]

2. 添加MS17010 [√]

3. 添加OXID [√]

4. 添加简单目录扫描 (将在新的工具中实现,gt主要进行资产探测)

5. 更灵活的端口模式 [√]

6. 更智能的扫描配置  [√]

7. 重构主要逻辑代码  [√]

8. 添加从文件中读取扫描目标  [√]

9. 添加常见服务指纹识别

   