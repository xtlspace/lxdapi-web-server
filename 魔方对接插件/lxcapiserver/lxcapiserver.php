<?php

use think\Db;

define('lxcapiserver_DEBUG', false);

function lxcapiserver_debug($message, $data = null)
{
    if (!lxcapiserver_DEBUG) return;
    $log = '[lxcapiserver-DEBUG] ' . $message;
    if ($data !== null) {
        $log .= ' | Data: ' . json_encode($data, JSON_UNESCAPED_UNICODE);
    }
    error_log($log);
}

function lxcapiserver_MetaData()
{
    return [
        'DisplayName' => '魔方财务-LxcAPI对接插件',
        'APIVersion'  => 'v2.4.5'
    ];
}

function lxcapiserver_ConfigOptions()
{
    return [
        'cpus' => [
            'type'        => 'text',
            'name'        => 'CPU核心数',
            'description' => 'CPU核心数量',
            'default'     => '1',
            'key'         => 'cpus',
        ],
        'memory' => [
            'type'        => 'text',
            'name'        => '内存',
            'description' => '内存大小，单位：MB',
            'default'     => '512',
            'key'         => 'memory',
        ],
        'disk' => [
            'type'        => 'text',
            'name'        => '硬盘',
            'description' => '硬盘大小，单位：MB',
            'default'     => '1024',
            'key'         => 'disk',
        ],
        'image' => [
            'type'        => 'text',
            'name'        => '镜像',
            'description' => '系统镜像名称',
            'default'     => 'alpine320',
            'key'         => 'image',
        ],
        'ingress' => [
            'type'        => 'text',
            'name'        => '入站带宽',
            'description' => '下载速度限制，单位：Mbit',
            'default'     => '100',
            'key'         => 'ingress',
        ],
        'egress' => [
            'type'        => 'text',
            'name'        => '出站带宽',
            'description' => '上传速度限制，单位：Mbit',
            'default'     => '100',
            'key'         => 'egress',
        ],
        'traffic_limit' => [
            'type'        => 'text',
            'name'        => '流量限制',
            'description' => '单位：GB，0表示不限制',
            'default'     => '100',
            'key'         => 'traffic_limit',
        ],
        'traffic_reset_cycle' => [
            'type'        => 'dropdown',
            'name'        => '流量重置周期',
            'description' => '按天或按月自动重置已用流量',
            'default'     => 'monthly',
            'key'         => 'traffic_reset_cycle',
            'options'     => ['monthly' => '按月', 'daily' => '按天'],
        ],
        'traffic_over_action' => [
            'type'        => 'dropdown',
            'name'        => '流量用尽后动作',
            'description' => '暂停容器，或限速后继续运行',
            'default'     => 'frozen',
            'key'         => 'traffic_over_action',
            'options'     => ['frozen' => '暂停容器', 'limited' => '限速'],
        ],
        'traffic_over_limitspeed' => [
            'type'        => 'text',
            'name'        => '流量用尽后限速',
            'description' => '仅动作为"限速"时生效，单位：Mbit',
            'default'     => '1',
            'key'         => 'traffic_over_limitspeed',
        ],
        'traffic_reset_price' => [
            'type'        => 'text',
            'name'        => '流量重置价格',
            'description' => '前台重置流量收费，单位：元/次，留空或0则不显示重置按钮',
            'default'     => '0',
            'key'         => 'traffic_reset_price',
        ],
        'ipv4_pool_limit' => [
            'type'        => 'text',
            'name'        => 'IPv4地址池限制',
            'description' => 'IPv4独立地址数量上限',
            'default'     => '0',
            'key'         => 'ipv4_pool_limit',
        ],
        'ipv4_mapping_limit' => [
            'type'        => 'text',
            'name'        => 'IPv4端口映射限制',
            'description' => 'IPv4端口转发规则上限',
            'default'     => '0',
            'key'         => 'ipv4_mapping_limit',
        ],
        'ipv6_pool_limit' => [
            'type'        => 'text',
            'name'        => 'IPv6地址池限制',
            'description' => 'IPv6独立地址数量上限',
            'default'     => '0',
            'key'         => 'ipv6_pool_limit',
        ],
        'ipv6_mapping_limit' => [
            'type'        => 'text',
            'name'        => 'IPv6端口映射限制',
            'description' => 'IPv6端口转发规则上限',
            'default'     => '0',
            'key'         => 'ipv6_mapping_limit',
        ],
        'reverse_proxy_limit' => [
            'type'        => 'text',
            'name'        => '反向代理限制',
            'description' => '反向代理域名数量上限',
            'default'     => '0',
            'key'         => 'reverse_proxy_limit',
        ],
        'cpu_allowance' => [
            'type'        => 'text',
            'name'        => 'CPU使用率限制',
            'description' => 'CPU占用百分比，单位：%',
            'default'     => '20',
            'key'         => 'cpu_allowance',
        ],
        'io_read' => [
            'type'        => 'text',
            'name'        => '磁盘读取限制',
            'description' => '单位：MB/s',
            'default'     => '100',
            'key'         => 'io_read',
        ],
        'io_write' => [
            'type'        => 'text',
            'name'        => '磁盘写入限制',
            'description' => '单位：MB/s',
            'default'     => '50',
            'key'         => 'io_write',
        ],
        'processes_limit' => [
            'type'        => 'text',
            'name'        => '最大进程数',
            'description' => '进程数量上限',
            'default'     => '512',
            'key'         => 'processes_limit',
        ],
        'allow_nesting' => [
            'type'        => 'dropdown',
            'name'        => '嵌套虚拟化',
            'description' => '支持Docker等虚拟化',
            'default'     => 'false',
            'key'         => 'allow_nesting',
            'options'     => ['true' => '启用', 'false' => '禁用'],
        ],
        'memory_swap' => [
            'type'        => 'dropdown',
            'name'        => 'Swap开关',
            'description' => '虚拟内存开关',
            'default'     => 'true',
            'key'         => 'memory_swap',
            'options'     => ['true' => '启用', 'false' => '禁用'],
        ],
        'privileged' => [
            'type'        => 'dropdown',
            'name'        => '特权模式',
            'description' => '特权容器开关',
            'default'     => 'false',
            'key'         => 'privileged',
            'options'     => ['true' => '启用', 'false' => '禁用'],
        ],
    ];
}

function lxcapiserver_ParseMemory($str)
{
    $str = trim($str);
    if (empty($str)) return 0;
    
    if (stripos($str, 'GB') !== false) {
        return intval($str) * 1024;
    } elseif (stripos($str, 'MB') !== false) {
        return intval($str);
    } else {
        return intval($str);
    }
}

function lxcapiserver_ParseBandwidth($str)
{
    $str = trim($str);
    if (empty($str)) return 0;
    
    if (stripos($str, 'Gbit') !== false) {
        return intval($str) * 1000;
    } elseif (stripos($str, 'Mbit') !== false) {
        return intval($str);
    } else {
        return intval($str);
    }
}

function lxcapiserver_ApiRequest($params, $endpoint, $data = [], $method = 'POST')
{
    $curl = curl_init();
    
    $protocol = 'https';
    $url = $protocol . '://' . $params['server_ip'] . ':' . $params['port'] . $endpoint;
    
    lxcapiserver_debug('API请求', [
        'url' => $url,
        'method' => $method
    ]);
    
    $curlOptions = [
        CURLOPT_URL            => $url,
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_ENCODING       => '',
        CURLOPT_MAXREDIRS      => 10,
        CURLOPT_TIMEOUT        => 30,
        CURLOPT_CONNECTTIMEOUT => 10,
        CURLOPT_FOLLOWLOCATION => true,
        CURLOPT_HTTP_VERSION   => CURL_HTTP_VERSION_1_1,
        CURLOPT_CUSTOMREQUEST  => $method,
        CURLOPT_HTTPHEADER     => [
            'X-API-Hash: ' . $params['accesshash'],
            'Content-Type: application/json',
        ],
    ];
    
    $curlOptions[CURLOPT_SSL_VERIFYPEER] = false;
    $curlOptions[CURLOPT_SSL_VERIFYHOST] = false;
    $curlOptions[CURLOPT_SSLVERSION] = CURL_SSLVERSION_TLSv1_2;
    
    if ($method === 'POST' || $method === 'PUT') {
        if (!empty($data)) {
            $curlOptions[CURLOPT_POSTFIELDS] = json_encode($data);
        }
    }
    
    curl_setopt_array($curl, $curlOptions);
    
    $response = curl_exec($curl);
    $errno = curl_errno($curl);
    $httpCode = curl_getinfo($curl, CURLINFO_HTTP_CODE);
    $curlError = curl_error($curl);
    
    curl_close($curl);
    
    lxcapiserver_debug('API响应', [
        'http_code' => $httpCode,
        'response_length' => strlen($response),
        'curl_errno' => $errno
    ]);
    
    if ($errno) {
        lxcapiserver_debug('CURL错误', [
            'errno' => $errno,
            'error' => $curlError
        ]);
        return null;
    }
    
    $decoded = json_decode($response, true);
    return $decoded;
}

function lxcapiserver_TestLink($params)
{
    lxcapiserver_debug('测试API连接', $params);
    
    $res = lxcapiserver_ApiRequest($params, '/api/system/containers', [], 'GET');
    
    lxcapiserver_debug('TestLink API响应', $res);
    
    if ($res === null) {
        return [
            'status' => 200,
            'data'   => [
                'server_status' => 0,
                'msg'           => '连接失败: 无法连接到服务器'
            ]
        ];
    }
    
    if (isset($res['code']) && $res['code'] == 200) {
        return [
            'status' => 200,
            'data'   => [
                'server_status' => 1,
                'msg'           => '连接成功'
            ]
        ];
    }
    
    return [
        'status' => 200,
        'data'   => [
            'server_status' => 0,
            'msg'           => '连接失败: ' . ($res['msg'] ?? '未知错误')
        ]
    ];
}

function lxcapiserver_CreateAccount($params)
{
    $containerName = is_array($params['domain']) ? $params['domain'][0] : $params['domain'];
    lxcapiserver_debug('创建容器', ['domain' => $containerName]);
    
    $configoptions = $params['configoptions'];
    
    $requestData = [
        'name' => $containerName,
        'image' => $configoptions['image'] ?? 'alpine320',
        'password' => $params['password'],
        'cpu' => (int)($configoptions['cpus'] ?? 1),
        'memory' => (int)($configoptions['memory'] ?? 512),
        'disk' => (int)($configoptions['disk'] ?? 1024),
        'ingress' => (int)($configoptions['ingress'] ?? 100),
        'egress' => (int)($configoptions['egress'] ?? 100),
        'traffic_limit' => (int)($configoptions['traffic_limit'] ?? 100),
        'traffic_reset_cycle' => $configoptions['traffic_reset_cycle'] ?? 'monthly',
        'traffic_over_action' => $configoptions['traffic_over_action'] ?? 'frozen',
        'traffic_over_limitspeed' => (int)($configoptions['traffic_over_limitspeed'] ?? 1),
        'allow_nesting' => ($configoptions['allow_nesting'] ?? 'false') === 'true',
        'memory_swap' => ($configoptions['memory_swap'] ?? 'true') === 'true',
        'privileged' => ($configoptions['privileged'] ?? 'false') === 'true',
        'cpu_allowance' => (int)($configoptions['cpu_allowance'] ?? 20),
        'io_read' => (int)($configoptions['io_read'] ?? 100),
        'io_write' => (int)($configoptions['io_write'] ?? 50),
        'processes_limit' => (int)($configoptions['processes_limit'] ?? 512),
        'ipv4_pool_limit' => (int)($configoptions['ipv4_pool_limit'] ?? 0),
        'ipv4_mapping_limit' => (int)($configoptions['ipv4_mapping_limit'] ?? 0),
        'ipv6_pool_limit' => (int)($configoptions['ipv6_pool_limit'] ?? 0),
        'ipv6_mapping_limit' => (int)($configoptions['ipv6_mapping_limit'] ?? 0),
        'reverse_proxy_limit' => (int)($configoptions['reverse_proxy_limit'] ?? 0),
    ];
    
    lxcapiserver_debug('创建请求数据', $requestData);
    
    $res = lxcapiserver_ApiRequest($params, '/api/system/containers', $requestData, 'POST');
    
    if ($res === null) {
        return ['status' => 'error', 'msg' => '请求失败'];
    }
    
    if (isset($res['code']) && $res['code'] == 200) {
        try {
            $update = [
                'domainstatus' => 'Active',
                'username'     => 'root',
                'dedicatedip'  => '请自行映射22端口',
            ];
            
            Db::name('host')->where('id', $params['hostid'])->update($update);
            lxcapiserver_debug('数据库更新成功', $update);
        } catch (\Exception $e) {
            return ['status' => 'error', 'msg' => '创建成功但同步数据失败: ' . $e->getMessage()];
        }
        
        return ['status' => 'success', 'msg' => $res['msg'] ?? '创建成功'];
    }
    
    return ['status' => 'error', 'msg' => $res['msg'] ?? '创建失败'];
}

function lxcapiserver_TerminateAccount($params)
{
    $containerName = is_array($params['domain']) ? $params['domain'][0] : $params['domain'];
    lxcapiserver_debug('删除容器', ['domain' => $containerName]);
    
    $endpoint = '/api/system/containers/' . urlencode($containerName);
    $res = lxcapiserver_ApiRequest($params, $endpoint, [], 'DELETE');
    
    if ($res === null) {
        return ['status' => 'error', 'msg' => '请求失败'];
    }
    
    if (isset($res['code']) && $res['code'] == 200) {
        return ['status' => 'success', 'msg' => $res['msg'] ?? '删除成功'];
    }
    
    return ['status' => 'error', 'msg' => $res['msg'] ?? '删除失败'];
}

function lxcapiserver_On($params)
{
    $containerName = is_array($params['domain']) ? $params['domain'][0] : $params['domain'];
    lxcapiserver_debug('启动容器', ['domain' => $containerName]);
    
    $endpoint = '/api/system/containers/' . urlencode($containerName) . '/action?action=start';
    $res = lxcapiserver_ApiRequest($params, $endpoint, [], 'POST');
    
    if ($res === null) {
        return ['status' => 'error', 'msg' => '请求失败'];
    }
    
    if (isset($res['code']) && $res['code'] == 200) {
        return ['status' => 'success', 'msg' => $res['msg'] ?? '启动成功'];
    }
    
    return ['status' => 'error', 'msg' => $res['msg'] ?? '启动失败'];
}

function lxcapiserver_Off($params)
{
    $containerName = is_array($params['domain']) ? $params['domain'][0] : $params['domain'];
    lxcapiserver_debug('停止容器', ['domain' => $containerName]);
    
    $endpoint = '/api/system/containers/' . urlencode($containerName) . '/action?action=stop';
    $res = lxcapiserver_ApiRequest($params, $endpoint, [], 'POST');
    
    if ($res === null) {
        return ['status' => 'error', 'msg' => '请求失败'];
    }
    
    if (isset($res['code']) && $res['code'] == 200) {
        return ['status' => 'success', 'msg' => $res['msg'] ?? '停止成功'];
    }
    
    return ['status' => 'error', 'msg' => $res['msg'] ?? '停止失败'];
}

function lxcapiserver_Reboot($params)
{
    $containerName = is_array($params['domain']) ? $params['domain'][0] : $params['domain'];
    lxcapiserver_debug('重启容器', ['domain' => $containerName]);
    
    $endpoint = '/api/system/containers/' . urlencode($containerName) . '/action?action=restart';
    $res = lxcapiserver_ApiRequest($params, $endpoint, [], 'POST');
    
    if ($res === null) {
        return ['status' => 'error', 'msg' => '请求失败'];
    }
    
    if (isset($res['code']) && $res['code'] == 200) {
        return ['status' => 'success', 'msg' => $res['msg'] ?? '重启成功'];
    }
    
    return ['status' => 'error', 'msg' => $res['msg'] ?? '重启失败'];
}

function lxcapiserver_SuspendAccount($params)
{
    $containerName = is_array($params['domain']) ? $params['domain'][0] : $params['domain'];
    lxcapiserver_debug('暂停容器', ['domain' => $containerName]);
    
    $endpoint = '/api/system/containers/' . urlencode($containerName) . '/action?action=pause';
    $res = lxcapiserver_ApiRequest($params, $endpoint, [], 'POST');
    
    if ($res === null) {
        return ['status' => 'error', 'msg' => '请求失败'];
    }
    
    if (isset($res['code']) && $res['code'] == 200) {
        return ['status' => 'success', 'msg' => $res['msg'] ?? '暂停成功'];
    }
    
    return ['status' => 'error', 'msg' => $res['msg'] ?? '暂停失败'];
}

function lxcapiserver_UnsuspendAccount($params)
{
    $containerName = is_array($params['domain']) ? $params['domain'][0] : $params['domain'];
    lxcapiserver_debug('恢复容器', ['domain' => $containerName]);
    
    $endpoint = '/api/system/containers/' . urlencode($containerName) . '/action?action=resume';
    $res = lxcapiserver_ApiRequest($params, $endpoint, [], 'POST');
    
    if ($res === null) {
        return ['status' => 'error', 'msg' => '请求失败'];
    }
    
    if (isset($res['code']) && $res['code'] == 200) {
        return ['status' => 'success', 'msg' => $res['msg'] ?? '恢复成功'];
    }
    
    return ['status' => 'error', 'msg' => $res['msg'] ?? '恢复失败'];
}

function lxcapiserver_Status($params)
{
    $containerName = is_array($params['domain']) ? $params['domain'][0] : $params['domain'];
    lxcapiserver_debug('查询状态', ['domain' => $containerName]);
    
    $endpoint = '/api/system/containers/' . urlencode($containerName);
    $res = lxcapiserver_ApiRequest($params, $endpoint, [], 'GET');
    
    if ($res === null) {
        return ['status' => 'error', 'msg' => '请求失败'];
    }
    
    if (isset($res['code']) && $res['code'] == 200 && isset($res['data']['status'])) {
        $containerStatus = $res['data']['status'];
        $result = ['status' => 'success'];

        $dbStatus = strtolower($res['data']['container']['status'] ?? '');
        if ($dbStatus === 'limited') {
            $result['data']['status'] = 'on';
            $result['data']['des'] = '已限速';
            return $result;
        }

        switch (strtoupper($containerStatus)) {
            case 'RUNNING':
                $result['data']['status'] = 'on';
                $result['data']['des'] = '运行中';
                break;
            case 'STOPPED':
                $result['data']['status'] = 'off';
                $result['data']['des'] = '已停止';
                break;
            case 'FROZEN':
                $result['data']['status'] = 'suspend';
                $result['data']['des'] = '已暂停';
                break;
            default:
                $result['data']['status'] = 'unknown';
                $result['data']['des'] = '未知状态';
                break;
        }
        
        return $result;
    }
    
    return ['status' => 'error', 'msg' => $res['msg'] ?? '查询失败'];
}

function lxcapiserver_Sync($params)
{
    $containerName = is_array($params['domain']) ? $params['domain'][0] : $params['domain'];
    lxcapiserver_debug('同步容器信息', ['domain' => $containerName]);
    
    $endpoint = '/api/system/containers/' . urlencode($containerName);
    $res = lxcapiserver_ApiRequest($params, $endpoint, [], 'GET');
    
    if ($res === null) {
        return ['status' => 'error', 'msg' => '请求失败'];
    }
    
    if (isset($res['code']) && $res['code'] == 200) {
        try {
            $update = [];
            
            if (isset($res['data']['status'])) {
                $containerStatus = strtoupper($res['data']['status']);
                if ($containerStatus === 'RUNNING') {
                    $update['domainstatus'] = 'Active';
                } elseif ($containerStatus === 'STOPPED') {
                    $update['domainstatus'] = 'Suspended';
                }
            }
            
            if (!empty($update)) {
                Db::name('host')->where('id', $params['hostid'])->update($update);
                lxcapiserver_debug('同步数据库状态成功', $update);
            }
            
            return ['status' => 'success', 'msg' => '状态同步成功'];
        } catch (\Exception $e) {
            return ['status' => 'error', 'msg' => '同步失败: ' . $e->getMessage()];
        }
    }
    
    return ['status' => 'error', 'msg' => $res['msg'] ?? '同步失败'];
}


function lxcapiserver_Reinstall($params)
{
    $containerName = is_array($params['domain']) ? $params['domain'][0] : $params['domain'];
    lxcapiserver_debug('重装系统', ['domain' => $containerName, 'reinstall_os' => $params['reinstall_os'] ?? 'null']);
    
    if (empty($params['reinstall_os'])) {
        return ['status' => 'error', 'msg' => '操作系统参数错误'];
    }
    
    $requestData = [
        'image' => $params['reinstall_os'],
        'password' => $params['password'],
    ];
    
    $endpoint = '/api/system/containers/' . urlencode($containerName) . '/action?action=reinstall';
    $res = lxcapiserver_ApiRequest($params, $endpoint, $requestData, 'POST');
    
    if ($res === null) {
        return ['status' => 'error', 'msg' => '请求失败'];
    }
    
    if (isset($res['code']) && $res['code'] == 200) {
        return ['status' => 'success', 'msg' => $res['msg'] ?? '重装成功'];
    }
    
    return ['status' => 'error', 'msg' => $res['msg'] ?? '重装失败'];
}

function lxcapiserver_AdminButton($params)
{
    if (!empty($params['domain'])) {
        return [
            'Sync' => '同步状态',
            'TrafficReset' => '重置流量',
        ];
    }
    return [];
}

function lxcapiserver_TrafficReset($params)
{
    $containerName = is_array($params['domain']) ? $params['domain'][0] : $params['domain'];
    lxcapiserver_debug('重置流量', ['domain' => $containerName]);
    
    $res = lxcapiserver_ApiRequest($params, '/api/system/traffic/reset?name=' . urlencode($containerName), [], 'POST');
    
    if ($res === null) {
        return ['status' => 'error', 'msg' => '请求失败'];
    }
    
    if (isset($res['code']) && $res['code'] == 200) {
        return ['status' => 'success', 'msg' => $res['msg'] ?? '流量重置成功'];
    }
    
    return ['status' => 'error', 'msg' => $res['msg'] ?? '流量重置失败'];
}

function lxcapiserver_EnsureTrafficResetTable()
{
    $prefix = \think\Db::getConfig('prefix');
    if (empty($prefix)) {
        $prefix = 'shd_';
    }
    $table = $prefix . 'lxc_traffic_reset';
    $exists = Db::query("SHOW TABLES LIKE '" . $table . "'");
    if (empty($exists)) {
        Db::execute("CREATE TABLE IF NOT EXISTS `" . $table . "` (
            `id` int(11) NOT NULL AUTO_INCREMENT,
            `uid` int(11) NOT NULL DEFAULT '0',
            `hostid` int(11) NOT NULL DEFAULT '0',
            `invoiceid` int(11) NOT NULL DEFAULT '0',
            `amount` decimal(10,2) NOT NULL DEFAULT '0.00',
            `status` varchar(20) NOT NULL DEFAULT 'pending',
            `create_time` int(11) NOT NULL DEFAULT '0',
            `paid_time` int(11) NOT NULL DEFAULT '0',
            `handle_time` int(11) NOT NULL DEFAULT '0',
            `remark` varchar(255) NOT NULL DEFAULT '',
            PRIMARY KEY (`id`),
            KEY `hostid` (`hostid`),
            KEY `invoiceid` (`invoiceid`)
        ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4");
    }
}

function lxcapiserver_DoTrafficResetByHost($hostid, $uid = 0)
{
    $hostid = intval($hostid);
    if ($hostid <= 0) {
        return ['status' => 'error', 'msg' => '主机参数错误'];
    }
    
    $hostinfo = Db::name('host')->where('id', $hostid)->find();
    if (empty($hostinfo)) {
        return ['status' => 'error', 'msg' => '主机不存在'];
    }
    if ($uid > 0 && intval($hostinfo['uid']) != intval($uid)) {
        return ['status' => 'error', 'msg' => '无权操作该主机'];
    }
    
    try {
        $hostModel = new \app\common\model\HostModel();
        $params = $hostModel->getProvisionParams($hostid);
    } catch (\Exception $e) {
        return ['status' => 'error', 'msg' => '获取主机配置失败: ' . $e->getMessage()];
    }
    if (empty($params) || empty($params['server_ip'])) {
        return ['status' => 'error', 'msg' => '获取主机配置失败'];
    }
    
    $res = lxcapiserver_TrafficReset($params);
    if (isset($res['status']) && $res['status'] == 'success') {
        try {
            Db::name('host')->where('id', $hostid)->update(['bwusage' => 0]);
        } catch (\Exception $e) {
            lxcapiserver_debug('清零流量用量失败', ['hostid' => $hostid, 'error' => $e->getMessage()]);
        }
        return ['status' => 'success', 'msg' => $res['msg'] ?? '流量重置成功'];
    }
    
    return ['status' => 'error', 'msg' => $res['msg'] ?? '流量重置失败'];
}

function lxcapiserver_AllowFunction()
{
    return [
        'client' => ['ResetTraffic'],
    ];
}

function lxcapiserver_ResetTraffic($params)
{
    $hostid = intval($params['hostid'] ?? 0);
    $uid = intval($params['uid'] ?? ($params['userid'] ?? 0));
    $containerName = is_array($params['domain']) ? $params['domain'][0] : ($params['domain'] ?? '');
    
    $price = floatval($params['configoptions']['traffic_reset_price'] ?? 0);
    if ($price <= 0) {
        return ['status' => 'error', 'msg' => '暂未开放流量重置功能'];
    }
    
    $hostinfo = Db::name('host')->where('id', $hostid)->where('uid', $uid)->find();
    if (empty($hostinfo)) {
        return ['status' => 'error', 'msg' => '主机不存在'];
    }
    if (empty($hostinfo['domain'])) {
        return ['status' => 'error', 'msg' => '容器信息不存在'];
    }
    if ($hostinfo['domainstatus'] != 'Active') {
        return ['status' => 'error', 'msg' => '当前状态不支持重置流量'];
    }
    
    lxcapiserver_EnsureTrafficResetTable();
    
    $existing = Db::name('lxc_traffic_reset')
        ->where('hostid', $hostid)
        ->where('uid', $uid)
        ->whereIn('status', ['pending', 'failed'])
        ->order('id', 'desc')
        ->find();
    if (!empty($existing)) {
        $invoiceInfo = Db::name('invoices')->where('id', $existing['invoiceid'])->find();
        $invStatus = $invoiceInfo['status'] ?? '';
        if ($invStatus == 'Unpaid') {
            return [
                'status'    => 200,
                'msg'       => '已有待支付的重置账单，请前往支付',
                'invoiceid' => intval($existing['invoiceid']),
                'data'      => ['invoiceid' => intval($existing['invoiceid'])],
            ];
        }
        if ($invStatus == 'Paid') {
            $resetResult = lxcapiserver_DoTrafficResetByHost($hostid, $uid);
            if (isset($resetResult['status']) && $resetResult['status'] == 'success') {
                Db::name('lxc_traffic_reset')->where('id', $existing['id'])->update([
                    'status'      => 'done',
                    'paid_time'   => intval($invoiceInfo['paid_time']),
                    'handle_time' => time(),
                    'remark'      => $resetResult['msg'],
                ]);
                return ['status' => 200, 'msg' => '流量已重置成功'];
            }
            Db::name('lxc_traffic_reset')->where('id', $existing['id'])->update([
                'handle_time' => time(),
                'remark'      => $resetResult['msg'],
            ]);
            return ['status' => 'error', 'msg' => '流量重置失败：' . $resetResult['msg']];
        }
    }
    
    $now = time();
    $invoiceData = [
        'uid'                 => $uid,
        'invoice_num'         => date('YmdHis') . mt_rand(1000, 9999),
        'create_time'         => $now,
        'update_time'         => $now,
        'due_time'            => $now + 86400 * 7,
        'paid_time'           => 0,
        'last_capture_attempt' => 0,
        'subtotal'            => $price,
        'credit'              => 0,
        'tax'                 => 0,
        'tax2'                => 0,
        'total'               => $price,
        'taxrate'             => 0,
        'taxrate2'            => 0,
        'status'              => 'Unpaid',
        'payment'             => '',
        'notes'               => '',
        'delete_time'         => 0,
        'due_email_times'     => 0,
        'type'                => 'product',
    ];
    $invoiceId = Db::name('invoices')->insertGetId($invoiceData);
    
    Db::name('invoice_items')->insert([
        'invoice_id'  => $invoiceId,
        'uid'         => $uid,
        'type'        => 'custom',
        'rel_id'      => $hostid,
        'description' => '容器流量重置（' . $containerName . '）',
        'amount'      => $price,
        'taxed'       => 0,
        'due_time'    => 0,
        'payment'     => '',
        'notes'       => '',
        'delete_time' => 0,
    ]);
    
    Db::name('lxc_traffic_reset')->insert([
        'uid'         => $uid,
        'hostid'      => $hostid,
        'invoiceid'   => $invoiceId,
        'amount'      => $price,
        'status'      => 'pending',
        'create_time' => $now,
        'paid_time'   => 0,
        'handle_time' => 0,
        'remark'      => '',
    ]);
    
    return [
        'status'    => 200,
        'msg'       => '账单已生成，请完成支付',
        'invoiceid' => intval($invoiceId),
        'data'      => ['invoiceid' => intval($invoiceId)],
    ];
}

function lxcapiserver_CrackPassword($params, $new_pass)
{
    $containerName = is_array($params['domain']) ? $params['domain'][0] : $params['domain'];
    lxcapiserver_debug('重置密码', ['domain' => $containerName]);
    
    $requestData = [
        'password' => $new_pass
    ];
    
    $endpoint = '/api/system/containers/' . urlencode($containerName) . '/action?action=reset-password';
    $res = lxcapiserver_ApiRequest($params, $endpoint, $requestData, 'POST');
    
    if ($res === null) {
        return ['status' => 'error', 'msg' => '请求失败'];
    }
    
    if (isset($res['code']) && $res['code'] == 200) {
        try {
            Db::name('host')->where('id', $params['hostid'])->update(['password' => $new_pass]);
        } catch (\Exception $e) {
            return ['status' => 'error', 'msg' => '密码重置成功但同步数据失败: ' . $e->getMessage()];
        }
        return ['status' => 'success', 'msg' => $res['msg'] ?? '密码重置成功'];
    }
    
    return ['status' => 'error', 'msg' => $res['msg'] ?? '密码重置失败'];
}

function lxcapiserver_vnc($params)
{
    $containerName = is_array($params['domain']) ? $params['domain'][0] : $params['domain'];
    lxcapiserver_debug('VNC控制台', ['domain' => $containerName]);
    
    $requestData = ['hostname' => $containerName];
    $res = lxcapiserver_ApiRequest($params, '/api/system/console/create-token', $requestData, 'POST');
    
    if ($res === null) {
        return ['status' => 'error', 'msg' => '请求失败'];
    }
    
    if (isset($res['code']) && $res['code'] == 200 && isset($res['data']['token'])) {
        $consoleUrl = 'https://' . $params['server_ip'] . ':' . $params['port'] . '/console?token=' . $res['data']['token'];
        
        return [
            'status' => 'success',
            'url' => $consoleUrl
        ];
    }
    
    return ['status' => 'error', 'msg' => $res['msg'] ?? 'VNC连接失败'];
}

function lxcapiserver_ClientArea($params)
{
    return [
        'info' => ['name' => '容器信息'],
    ];
}

function lxcapiserver_ClientAreaOutput($params, $key)
{
    lxcapiserver_debug('ClientAreaOutput调用', ['key' => $key]);
    
    if ($key == 'info') {
        $containerName = is_array($params['domain']) ? $params['domain'][0] : $params['domain'];
        $hostid = intval($params['hostid'] ?? 0);
        $uid = intval($params['uid'] ?? ($params['userid'] ?? 0));
        
        $resetPrice = floatval($params['configoptions']['traffic_reset_price'] ?? 0);
        $trafficResetEnabled = $resetPrice > 0;
        $resetNotice = '';
        $resetNoticeType = '';
        $hasPending = false;
        $pendingInvoiceId = 0;
        
        if ($trafficResetEnabled && $hostid > 0 && $uid > 0) {
            try {
                lxcapiserver_EnsureTrafficResetTable();
                $pendingList = Db::name('lxc_traffic_reset')
                    ->where('hostid', $hostid)
                    ->where('uid', $uid)
                    ->whereIn('status', ['pending', 'failed'])
                    ->order('id', 'desc')
                    ->limit(20)
                    ->select()
                    ->toArray();
                foreach ($pendingList as $pending) {
                    $invoiceInfo = Db::name('invoices')->where('id', $pending['invoiceid'])->find();
                    if (empty($invoiceInfo)) {
                        continue;
                    }
                    if ($invoiceInfo['status'] == 'Paid') {
                        $resetResult = lxcapiserver_DoTrafficResetByHost($hostid, $uid);
                        if (isset($resetResult['status']) && $resetResult['status'] == 'success') {
                            Db::name('lxc_traffic_reset')->where('id', $pending['id'])->update([
                                'status'      => 'done',
                                'paid_time'   => intval($invoiceInfo['paid_time']),
                                'handle_time' => time(),
                                'remark'      => $resetResult['msg'],
                            ]);
                            $resetNotice = '流量已重置成功';
                            $resetNoticeType = 'success';
                        } else {
                            Db::name('lxc_traffic_reset')->where('id', $pending['id'])->update([
                                'handle_time' => time(),
                                'remark'      => $resetResult['msg'],
                            ]);
                            $resetNotice = '流量重置失败：' . $resetResult['msg'];
                            $resetNoticeType = 'danger';
                        }
                    } elseif ($invoiceInfo['status'] == 'Unpaid') {
                        $hasPending = true;
                        $pendingInvoiceId = intval($pending['invoiceid']);
                        break;
                    }
                }
            } catch (\Exception $e) {
                lxcapiserver_debug('流量重置兜底处理异常', ['error' => $e->getMessage()]);
            }
        }
        
        $endpoint = '/api/system/containers/' . urlencode($containerName) . '/credential';
        $res = lxcapiserver_ApiRequest($params, $endpoint, [], 'GET');
        
        $jumpUrl = '';
        $iframeUrl = '';
        $accessCode = '';
        $errorMsg = '';
        
        if (isset($res['code']) && $res['code'] == 200 && isset($res['data'])) {
            $accessCode = $res['data']['access_code'] ?? '';
            $protocol = 'https';
            $baseUrl = $protocol . '://' . $params['server_ip'] . ':' . $params['port'];
            $jumpUrl = $baseUrl . '/container/dashboard?hash=' . $accessCode;
            $iframeUrl = $baseUrl . '/container/dashboard/lite?hash=' . $accessCode;
        } else {
            $errorMsg = $res['msg'] ?? '获取访问码失败';
        }
        
        return [
            'template' => 'templates/info.html',
            'vars' => [
                'container_name' => $containerName,
                'server_ip' => $params['server_ip'],
                'server_port' => $params['port'],
                'jump_url' => $jumpUrl,
                'iframe_url' => $iframeUrl,
                'access_code' => $accessCode,
                'error_msg' => $errorMsg,
                'traffic_reset_enabled' => $trafficResetEnabled,
                'reset_price' => number_format($resetPrice, 2),
                'reset_notice' => $resetNotice,
                'reset_notice_type' => $resetNoticeType,
                'has_pending' => $hasPending,
                'pending_invoiceid' => $pendingInvoiceId,
            ]
        ];
    }
    
    return '';
}


function lxcapiserver_ChangePackage($params)
{
    $containerName = is_array($params['domain']) ? $params['domain'][0] : $params['domain'];
    lxcapiserver_debug('升级配置', ['domain' => $containerName]);
    
    $configoptions = $params['configoptions'];
    
    $requestData = [
        'cpu'                => (int)($configoptions['cpus'] ?? 0) ?: null,
        'memory'             => (int)($configoptions['memory'] ?? 0) ?: null,
        'disk'               => (int)($configoptions['disk'] ?? 0) ?: null,
        'ingress'            => (int)($configoptions['ingress'] ?? 0) ?: null,
        'egress'             => (int)($configoptions['egress'] ?? 0) ?: null,
        'traffic_limit'      => (int)($configoptions['traffic_limit'] ?? 0) ?: null,
        'traffic_reset_cycle' => in_array($configoptions['traffic_reset_cycle'] ?? '', ['monthly', 'daily'], true)
            ? $configoptions['traffic_reset_cycle'] : null,
        'traffic_over_action' => in_array($configoptions['traffic_over_action'] ?? '', ['frozen', 'limited'], true)
            ? $configoptions['traffic_over_action'] : null,
        'traffic_over_limitspeed' => isset($configoptions['traffic_over_limitspeed'])
            ? (int)$configoptions['traffic_over_limitspeed'] : null,
        'cpu_allowance'      => (int)($configoptions['cpu_allowance'] ?? 0) ?: null,
        'io_read'            => (int)($configoptions['io_read'] ?? 0) ?: null,
        'io_write'           => (int)($configoptions['io_write'] ?? 0) ?: null,
        'processes_limit'    => (int)($configoptions['processes_limit'] ?? 0) ?: null,
        'ipv4_pool_limit'    => (int)($configoptions['ipv4_pool_limit'] ?? 0) ?: null,
        'ipv4_mapping_limit' => (int)($configoptions['ipv4_mapping_limit'] ?? 0) ?: null,
        'ipv6_pool_limit'    => (int)($configoptions['ipv6_pool_limit'] ?? 0) ?: null,
        'ipv6_mapping_limit' => (int)($configoptions['ipv6_mapping_limit'] ?? 0) ?: null,
        'reverse_proxy_limit'=> (int)($configoptions['reverse_proxy_limit'] ?? 0) ?: null,
    ];
    
    if (isset($configoptions['allow_nesting'])) {
        $requestData['allow_nesting'] = $configoptions['allow_nesting'] === 'true';
    }
    if (isset($configoptions['memory_swap'])) {
        $requestData['memory_swap'] = $configoptions['memory_swap'] === 'true';
    }
    if (isset($configoptions['privileged'])) {
        $requestData['privileged'] = $configoptions['privileged'] === 'true';
    }
    
    $requestData = array_filter($requestData, function($v) { return $v !== null; });
    
    lxcapiserver_debug('升级请求数据', $requestData);
    
    $endpoint = '/api/system/containers/' . urlencode($containerName) . '/config';
    $res = lxcapiserver_ApiRequest($params, $endpoint, $requestData, 'PUT');
    
    if ($res === null) {
        return ['status' => 'error', 'msg' => '请求失败'];
    }
    
    if (isset($res['code']) && $res['code'] == 200) {
        return ['status' => 'success', 'msg' => '配置升级成功'];
    }
    
    return ['status' => 'error', 'msg' => $res['msg'] ?? '升级失败'];
}

