
 <h1 class="curproject-name"> AnsibleAPI </h1> 
 ansible 扩展restful API


# 主机接口

## 主机检查
<a id=主机检查> </a>
### 基本信息

**Path：** /api/v1/ansible/host_check

**Method：** POST

**接口描述：**


### 请求参数
**Headers**

| 参数名称  | 参数值  |  是否必须 | 示例  | 备注  |
| ------------ | ------------ | ------------ | ------------ | ------------ |
| Content-Type  |  application/json | 是  |   |   |
**Body**

<table>
  <thead class="ant-table-thead">
    <tr>
      <th key=name>名称</th><th key=type>类型</th><th key=required>是否必须</th><th key=default>默认值</th><th key=desc>备注</th><th key=sub>其他信息</th>
    </tr>
  </thead><tbody className="ant-table-tbody"><tr key=0><td key=0><span style="padding-left: 0px"><span style="color: #8c8a8a"></span> </span></td><td key=1><span></span></td><td key=2>非必须</td><td key=3></td><td key=4><span></span></td><td key=5></td></tr>
               </tbody>
              </table>
            
### 返回数据

<table>
  <thead class="ant-table-thead">
    <tr>
      <th key=name>名称</th><th key=type>类型</th><th key=required>是否必须</th><th key=default>默认值</th><th key=desc>备注</th><th key=sub>其他信息</th>
    </tr>
  </thead><tbody className="ant-table-tbody"><tr key=0><td key=0><span style="padding-left: 0px"><span style="color: #8c8a8a"></span> </span></td><td key=1><span></span></td><td key=2>非必须</td><td key=3></td><td key=4><span></span></td><td key=5></td></tr>
               </tbody>
              </table>
            
## 添加主机
<a id=添加主机> </a>
### 基本信息

**Path：** /api/v1/ansible/hosts

**Method：** POST

**接口描述：**


### 请求参数
**Headers**

| 参数名称  | 参数值  |  是否必须 | 示例  | 备注  |
| ------------ | ------------ | ------------ | ------------ | ------------ |
| Content-Type  |  application/json | 是  |   |   |
**Body**

<table>
  <thead class="ant-table-thead">
    <tr>
      <th key=name>名称</th><th key=type>类型</th><th key=required>是否必须</th><th key=default>默认值</th><th key=desc>备注</th><th key=sub>其他信息</th>
    </tr>
  </thead><tbody className="ant-table-tbody"><tr key=0><td key=0><span style="padding-left: 0px"><span style="color: #8c8a8a"></span> </span></td><td key=1><span></span></td><td key=2>非必须</td><td key=3></td><td key=4><span></span></td><td key=5></td></tr>
               </tbody>
              </table>
            
### 返回数据

<table>
  <thead class="ant-table-thead">
    <tr>
      <th key=name>名称</th><th key=type>类型</th><th key=required>是否必须</th><th key=default>默认值</th><th key=desc>备注</th><th key=sub>其他信息</th>
    </tr>
  </thead><tbody className="ant-table-tbody"><tr key=0><td key=0><span style="padding-left: 0px"><span style="color: #8c8a8a"></span> </span></td><td key=1><span></span></td><td key=2>非必须</td><td key=3></td><td key=4><span></span></td><td key=5></td></tr>
               </tbody>
              </table>
            
# Playbook接口

## Playbook上传
<a id=Playbook上传> </a>
### 基本信息

**Path：** /api/v1/ansible/playbooks/{name}/{version}

**Method：** POST

**接口描述：**


### 请求参数
**路径参数**
| 参数名称 | 示例  | 备注  |
| ------------ | ------------ | ------------ | ------------ | ------------ |
| name |  redis_exporter |   |
| version |  v0.0.1 |   |

### 返回数据

```javascript
{
   "took": "35.963088ms"
}
```
## Playbook列表
<a id=Playbook列表> </a>
### 基本信息

**Path：** /api/v1/ansible/playbooks/{name}

**Method：** GET

**接口描述：**


### 请求参数
**路径参数**
| 参数名称 | 示例  | 备注  |
| ------------ | ------------ | ------------ | ------------ | ------------ |
| name |  redis_exporter |   |

### 返回数据

```javascript
[
   {
      "name": "redis_exporter",
      "version": "v0.0.1",
      "link": "/api/v1/ansible/playbooks/redis_exporter/v0.0.1"
   }
]
```
## Playbook删除
<a id=Playbook删除> </a>
### 基本信息

**Path：** /api/v1/ansible/playbooks/{name}/{version}

**Method：** DELETE

**接口描述：**


### 请求参数
**Headers**

| 参数名称  | 参数值  |  是否必须 | 示例  | 备注  |
| ------------ | ------------ | ------------ | ------------ | ------------ |
| Content-Type  |  application/x-www-form-urlencoded | 是  |   |   |
**路径参数**
| 参数名称 | 示例  | 备注  |
| ------------ | ------------ | ------------ | ------------ | ------------ |
| name |   |   |
| version |   |   |

## Playbook运行
<a id=Playbook运行> </a>
### 基本信息

**Path：** /api/v1/ansible/play/{name}/{version}

**Method：** POST

**接口描述：**


### 请求参数
**Headers**

| 参数名称  | 参数值  |  是否必须 | 示例  | 备注  |
| ------------ | ------------ | ------------ | ------------ | ------------ |
| Content-Type  |  application/json | 是  |   |   |
**路径参数**
| 参数名称 | 示例  | 备注  |
| ------------ | ------------ | ------------ | ------------ | ------------ |
| name |  redis_exporter |   |
| version |  v0.0.1 |   |
**Body**

```javascript
{
   "params": {
      "listen_port": "9121",
      "options": "-redis.addr 10.138.16.188:6379 -redis.password 123123 -separator ,,  -redis.alias=CMP-prod"
   },
   "hosts": [
      "10.138.16.188"
   ],
   "register": {
      "listenPort": 9121,
      "labels": {
         "env": "prod",
         "project": "CMP"
      }
   }
}
```
### 返回数据

```javascript
{"result":{"job":"redis_exporter","type":"PLAY","name":"10.138.16.188"}}
{"result":{"job":"redis_exporter","type":"TASK","name":"Gathering Facts"}}
{"result":{"job":"redis_exporter","type":"HOST","host":"10.138.16.188","step":1,"name":"Gathering Facts","status":"ok","progress":6}}
{"result":{"job":"redis_exporter","type":"TASK","name":"create exporter group"}}
{"result":{"job":"redis_exporter","type":"HOST","host":"10.138.16.188","step":2,"name":"create exporter group","status":"ok","progress":13}}
{"result":{"job":"redis_exporter","type":"TASK","name":"create exporter user"}}
{"result":{"job":"redis_exporter","type":"HOST","host":"10.138.16.188","step":3,"name":"create exporter user","status":"ok","progress":20}}
{"result":{"job":"redis_exporter","type":"TASK","name":"redis_exporter service file"}}
{"result":{"job":"redis_exporter","type":"HOST","host":"10.138.16.188","step":4,"name":"redis_exporter service file","status":"changed","progress":26}}
{"result":{"job":"redis_exporter","type":"TASK","name":"set running script args for listen port"}}
{"result":{"job":"redis_exporter","type":"HOST","host":"10.138.16.188","step":5,"name":"set running script args for listen port","status":"changed","progress":33}}
{"result":{"job":"redis_exporter","type":"TASK","name":"set running script args"}}
{"result":{"job":"redis_exporter","type":"HOST","host":"10.138.16.188","step":6,"name":"set running script args","status":"changed","progress":40}}
{"result":{"job":"redis_exporter","type":"TASK","name":"create Downloads dir"}}
{"result":{"job":"redis_exporter","type":"HOST","host":"10.138.16.188","step":7,"name":"create Downloads dir","status":"ok","progress":46}}
{"result":{"job":"redis_exporter","type":"TASK","name":"download the redis_exporter pkg"}}
{"result":{"job":"redis_exporter","type":"HOST","host":"10.138.16.188","step":8,"name":"download the redis_exporter pkg","status":"ok","progress":53}}
{"result":{"job":"redis_exporter","type":"TASK","name":"unzip redis_exporter pkg"}}
{"result":{"job":"redis_exporter","type":"HOST","host":"10.138.16.188","step":9,"name":"unzip redis_exporter pkg","status":"ok","progress":60}}
{"result":{"job":"redis_exporter","type":"TASK","name":"move redis_exporter"}}
{"result":{"job":"redis_exporter","type":"HOST","host":"10.138.16.188","step":10,"name":"move redis_exporter","status":"changed","progress":66}}
{"result":{"job":"redis_exporter","type":"TASK","name":"enable redis_exporter_9121"}}
{"result":{"job":"redis_exporter","type":"HOST","host":"10.138.16.188","step":11,"name":"enable redis_exporter_9121","status":"ok","progress":73}}
{"result":{"job":"redis_exporter","type":"TASK","name":"enable redis_exporter_9121"}}
{"result":{"job":"redis_exporter","type":"HOST","host":"10.138.16.188","step":12,"name":"enable redis_exporter_9121","status":"skipping","progress":80}}
{"result":{"job":"redis_exporter","type":"TASK","name":"enable redis_exporter_9121"}}
{"result":{"job":"redis_exporter","type":"HOST","host":"10.138.16.188","step":13,"name":"enable redis_exporter_9121","status":"ok","progress":86}}
{"result":{"job":"redis_exporter","type":"TASK","name":"start redis_exporter_9121"}}
{"result":{"job":"redis_exporter","type":"HOST","host":"10.138.16.188","step":14,"name":"start redis_exporter_9121","status":"changed","progress":93}}
{"result":{"job":"redis_exporter","type":"TASK","name":"configure firewall"}}
{"result":{"job":"redis_exporter","type":"HOST","host":"10.138.16.188","step":15,"name":"configure firewall","status":"changed","progress":99}}
{"result":{"job":"redis_exporter","type":"TASK","name":"reload firewall"}}
{"result":{"job":"redis_exporter","type":"HOST","host":"10.138.16.188","step":16,"name":"reload firewall","status":"changed","progress":99}}
{"result":{"job":"redis_exporter","type":"PLAY","name":"RECAP"}}
{"result":{"job":"redis_exporter","type":"RECAP","host":"10.138.16.188","status":"ok","ok":15,"changed":7,"progress":100}}
```