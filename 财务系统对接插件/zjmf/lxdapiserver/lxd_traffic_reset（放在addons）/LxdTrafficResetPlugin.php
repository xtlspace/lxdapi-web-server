<?php

namespace addons\lxd_traffic_reset;

use think\Db;

class LxdTrafficResetPlugin extends \app\admin\lib\Plugin
{
    public $info = [
        'name'        => 'LxdTrafficReset',
        'title'       => 'LXD流量重置支付回调',
        'description' => '监听账单支付成功事件，自动执行LXD容器流量重置（配合 lxdapiserver 服务器模块使用）',
        'status'      => 1,
        'author'      => 'xkatld',
        'version'     => '1.0.0',
        'module'      => 'addons',
    ];

    public $hasAdmin = 0;

    public function install()
    {
        $this->ensureTable();
        return true;
    }

    public function uninstall()
    {
        return true;
    }

    public function invoicePaid($params)
    {
        $invoiceid = intval($params['invoiceid'] ?? 0);
        if ($invoiceid <= 0) {
            return;
        }

        try {
            $pending = Db::name('lxd_traffic_reset')
                ->where('invoiceid', $invoiceid)
                ->where('status', 'pending')
                ->find();
            if (empty($pending)) {
                return;
            }

            $invoiceInfo = Db::name('invoices')->where('id', $invoiceid)->find();
            if (empty($invoiceInfo) || $invoiceInfo['status'] != 'Paid') {
                return;
            }
            if (intval($invoiceInfo['uid']) != intval($pending['uid'])) {
                return;
            }

            $moduleFile = WEB_ROOT . 'plugins/servers/lxdapiserver/lxdapiserver.php';
            if (!file_exists($moduleFile)) {
                Db::name('lxd_traffic_reset')->where('id', $pending['id'])->update([
                    'status'      => 'failed',
                    'paid_time'   => intval($invoiceInfo['paid_time']),
                    'handle_time' => time(),
                    'remark'      => '服务器模块文件不存在',
                ]);
                return;
            }

            require_once $moduleFile;

            $result = \lxdapiserver_DoTrafficResetByHost($pending['hostid'], $pending['uid']);
            if (isset($result['status']) && $result['status'] == 'success') {
                Db::name('lxd_traffic_reset')->where('id', $pending['id'])->update([
                    'status'      => 'done',
                    'paid_time'   => intval($invoiceInfo['paid_time']),
                    'handle_time' => time(),
                    'remark'      => $result['msg'],
                ]);
            } else {
                Db::name('lxd_traffic_reset')->where('id', $pending['id'])->update([
                    'status'      => 'failed',
                    'paid_time'   => intval($invoiceInfo['paid_time']),
                    'handle_time' => time(),
                    'remark'      => $result['msg'],
                ]);
            }
        } catch (\Exception $e) {
            try {
                Db::name('lxd_traffic_reset')->where('invoiceid', $invoiceid)->where('status', 'pending')->update([
                    'status'      => 'failed',
                    'handle_time' => time(),
                    'remark'      => mb_substr($e->getMessage(), 0, 200),
                ]);
            } catch (\Exception $ignored) {
            }
        }
    }

    private function ensureTable()
    {
        $prefix = \think\Db::getConfig('prefix');
        if (empty($prefix)) {
            $prefix = 'shd_';
        }
        $table = $prefix . 'lxd_traffic_reset';
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
}
