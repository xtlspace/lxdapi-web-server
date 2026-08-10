<?php
namespace addons\wjssk_push;

class WjsskPushPlugin extends \app\admin\lib\Plugin
{
    public $info = ["name" => "WjsskPush", "title" => "魔方消息推送插件", "description" => "魔方消息推送插件，将集成多种除了邮箱以外推送方式", "status" => 1, "author" => "kid_jok", "version" => "1.0.0", "module" => "addons", "lang" => ["chinese" => "魔方消息推送插件", "chinese_tw" => "魔方消息推送插件", "english" => "Plug-in message push"]];
    public function install()
    {
        $DbConfig = \think\Db::getConfig();
        $LogTable = $DbConfig["prefix"] . "wjssk_push_log";
        $exist = \think\Db::query("show tables like \"" . $LogTable . "\"");
        if (!$exist) {
            \think\Db::query("CREATE TABLE `" . $LogTable . "`  (`id` int(11) NOT NULL AUTO_INCREMENT,`date_time` datetime(0) NULL DEFAULT NULL,`type` varchar(255) CHARACTER SET utf8 COLLATE utf8_general_ci NULL DEFAULT NULL,`content` longtext CHARACTER SET utf8 COLLATE utf8_general_ci NULL,`result` varchar(64) CHARACTER SET utf8 COLLATE utf8_general_ci NULL,PRIMARY KEY (`id`) USING BTREE) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8 COLLATE = utf8_general_ci ROW_FORMAT = Dynamic;");
        }
        return true;
    }
    public function uninstall()
    {
        return true;
    }
    public function errorFilter($type, $name, $config, $params, $msg = "")
    {
        $host = \think\Db::name("host")->where("id", $params["hostid"])->find();
        $pushMsg = "产品" . $name . "<code>失败</code>通知\n";
        $pushMsg .= "产品名称：[" . $params["hostid"] . "]-" . $params["name"] . "\n";
        $pushMsg .= "产品IP：[<code>" . $host["dedicatedip"] . "</code>]\n\n";
        $pushMsg .= "失败原因：" . $msg . "\n\n";
        $pushMsg .= "客户姓名：[" . $params["user_info"]["id"] . "]-" . $params["user_info"]["username"] . "\n";
        $pushMsg .= "操作时间：" . date("Y-m-d H:i:s") . "\n\n";
        $pushMsg .= "通知来自【<a href='" . configuration("system_url") . "'>" . configuration("company_name") . "</a>】";
        $this->pushByTg($pushMsg, $config, $type);
    }
    public function ticketUserReply($params)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        $priority = ["high" => "高", "medium" => "中", "low" => "低"];
        $ticket = \think\Db::name("ticket")->where("id", $params["ticketid"])->find();
        $host = \think\Db::name("host")->where("id", $ticket["host_id"])->find();
        $user = \think\Db::name("clients")->where("id", $params["uid"])->find();
        $product = \think\Db::name("products")->where("id", $host["productid"])->find();
        if ($tg_push && $config["ticket_user_reply"] == 1) {
            $pushMsg = "用户回复工单通知\n";
            $pushMsg .= "工单信息：" . $params["ticketid"] . "-" . $params["status_title"] . "\n";
            $pushMsg .= "工单部门：" . $params["dptname"] . "\n";
            $pushMsg .= "工单标题：" . $params["title"] . "\n";
            $pushMsg .= "工单内容：------\n" . $params["content"] . "\n------------------\n";
            $pushMsg .= "客户姓名：[" . $user["id"] . "]-" . $user["username"] . "\n";
            $pushMsg .= "工单相关产品：[" . $host["id"] . "]-" . $product["name"] . "\n";
            $pushMsg .= "工单优先级：" . $priority[$params["priority"]] . "\n";
            $pushMsg .= "操作时间：" . date("Y-m-d H:i:s") . "\n\n";
            $pushMsg .= "通知来自【<a href='" . configuration("system_url") . "'>" . configuration("company_name") . "</a>】";
            $this->pushByTg($pushMsg, $config, "ticket_user_reply");
        }
    }
    public function ticketOpen($params)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        $priority = ["high" => "高", "medium" => "中", "low" => "低"];
        $host = \think\Db::name("host")->where("id", $params["hostid"])->find();
        $user = \think\Db::name("clients")->where("id", $params["uid"])->find();
        $product = \think\Db::name("products")->where("id", $host["productid"])->find();
        if ($tg_push && $config["ticket_open"] == 1) {
            $pushMsg = "用户创建工单通知\n";
            $pushMsg .= "工单ID：" . $params["ticketid"] . "\n";
            $pushMsg .= "工单部门：" . $params["dptname"] . "\n";
            $pushMsg .= "工单标题：" . $params["title"] . "\n";
            $pushMsg .= "工单内容：------\n" . $params["content"] . "\n------------------\n";
            $pushMsg .= "客户姓名：[" . $user["id"] . "]-" . $user["username"] . "\n";
            $pushMsg .= "工单相关产品：[" . $params["hostid"] . "]-" . $product["name"] . "\n";
            $pushMsg .= "工单优先级：" . $priority[$params["priority"]] . "\n";
            $pushMsg .= "操作时间：" . date("Y-m-d H:i:s") . "\n\n";
            $pushMsg .= "通知来自【<a href='" . configuration("system_url") . "'>" . configuration("company_name") . "</a>】";
            $this->pushByTg($pushMsg, $config, "ticket_open");
        }
    }
    public function afterModuleCreate($params)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        if ($tg_push && $config["after_module_create"] == 1) {
            $params = $params["params"];
            $host = \think\Db::name("host")->where("id", $params["hostid"])->find();
            $pushMsg = "产品开通成功通知\n";
            $pushMsg .= "产品名称：[" . $params["hostid"] . "]-" . $params["name"] . "\n";
            $pushMsg .= "产品IP：[<code>" . $host["dedicatedip"] . "</code>]\n";
            $pushMsg .= "登录信息：<code>" . $host["username"] . "</code>(<code>" . $host["password"] . "</code>)\n";
            $pushMsg .= "到期时间：" . date("Y-m-d H:i:s", $host["nextduedate"]) . "\n\n";
            $pushMsg .= "客户姓名：[" . $params["user_info"]["id"] . "]-" . $params["user_info"]["username"] . "\n";
            $pushMsg .= "操作时间：" . date("Y-m-d H:i:s") . "\n\n";
            $pushMsg .= "通知来自【<a href='" . configuration("system_url") . "'>" . configuration("company_name") . "</a>】";
            $this->pushByTg($pushMsg, $config, "after_module_create");
        }
    }
    public function afterModuleCreateFailed($params)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        if ($tg_push && $config["after_module_create"] == 1) {
            $msg = $params["msg"];
            $params = $params["params"];
            $this->errorFilter("after_module_create_failed", "开通", $config, $params, $msg);
        }
    }
    public function afterModuleCrackPassword($params)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        if ($tg_push && $config["after_module_crack_password"] == 1) {
            $host = \think\Db::name("host")->where("id", $params["hostid"])->find();
            $user = \think\Db::name("clients")->where("id", $host["uid"])->find();
            $product = \think\Db::name("products")->where("id", $host["productid"])->find();
            $pushMsg = "产品重置密码成功通知\n";
            $pushMsg .= "产品名称：[" . $params["hostid"] . "]-" . $product["name"] . "\n";
            $pushMsg .= "产品IP：[<code>" . $host["dedicatedip"] . "</code>]\n";
            $pushMsg .= "旧密码：<code>" . $params["oldpassword"] . "</code>\n";
            $pushMsg .= "新密码：<code>" . $params["newspassword"] . "</code>\n\n";
            $pushMsg .= "客户姓名：[" . $user["id"] . "]-" . $user["username"] . "\n";
            $pushMsg .= "操作时间：" . date("Y-m-d H:i:s") . "\n\n";
            $pushMsg .= "通知来自【<a href='" . configuration("system_url") . "'>" . configuration("company_name") . "</a>】";
            $this->pushByTg($pushMsg, $config, "after_module_crack_password");
        }
    }
    public function afterModuleCrackPasswordFailed($params)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        if ($tg_push && $config["after_module_crack_password"] == 1) {
            $msg = $params["msg"];
            $params = $params["params"];
            $this->errorFilter("after_module_crack_password_failed", "重置密码", $config, $params, $msg);
        }
    }
    public function afterModuleOn($params)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        if ($tg_push && $config["after_module_on"] == 1) {
            $params = $params["params"];
            $host = \think\Db::name("host")->where("id", $params["hostid"])->find();
            $pushMsg = "产品开机成功通知\n";
            $pushMsg .= "产品名称：[" . $params["hostid"] . "]-" . $params["name"] . "\n";
            $pushMsg .= "产品IP：[<code>" . $host["dedicatedip"] . "</code>]\n\n";
            $pushMsg .= "客户姓名：[" . $params["user_info"]["id"] . "]-" . $params["user_info"]["username"] . "\n";
            $pushMsg .= "操作时间：" . date("Y-m-d H:i:s") . "\n\n";
            $pushMsg .= "通知来自【<a href='" . configuration("system_url") . "'>" . configuration("company_name") . "</a>】";
            $this->pushByTg($pushMsg, $config, "after_module_on");
        }
    }
    public function afterModuleOnFailed($params)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        if ($tg_push && $config["after_module_on"] == 1) {
            $msg = $params["msg"];
            $params = $params["params"];
            $this->errorFilter("after_module_on_failed", "开机", $config, $params, $msg);
        }
    }
    public function afterModuleOff($params)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        if ($tg_push && $config["after_module_off"] == 1) {
            $params = $params["params"];
            $host = \think\Db::name("host")->where("id", $params["hostid"])->find();
            $pushMsg = "产品关机成功通知\n";
            $pushMsg .= "产品名称：[" . $params["hostid"] . "]-" . $params["name"] . "\n";
            $pushMsg .= "产品IP：[<code>" . $host["dedicatedip"] . "</code>]\n\n";
            $pushMsg .= "客户姓名：[" . $params["user_info"]["id"] . "]-" . $params["user_info"]["username"] . "\n";
            $pushMsg .= "操作时间：" . date("Y-m-d H:i:s") . "\n\n";
            $pushMsg .= "通知来自【<a href='" . configuration("system_url") . "'>" . configuration("company_name") . "</a>】";
            $this->pushByTg($pushMsg, $config, "after_module_off");
        }
    }
    public function afterModuleOffFailed($params)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        if ($tg_push && $config["after_module_off"] == 1) {
            $msg = $params["msg"];
            $params = $params["params"];
            $this->errorFilter("after_module_off_failed", "关机", $config, $params, $msg);
        }
    }
    public function afterModuleHardOff($params)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        if ($tg_push && $config["after_module_hard_off"] == 1) {
            $params = $params["params"];
            $host = \think\Db::name("host")->where("id", $params["hostid"])->find();
            $pushMsg = "产品硬关机成功通知\n";
            $pushMsg .= "产品名称：[" . $params["hostid"] . "]-" . $params["name"] . "\n";
            $pushMsg .= "产品IP：[<code>" . $host["dedicatedip"] . "</code>]\n\n";
            $pushMsg .= "客户姓名：[" . $params["user_info"]["id"] . "]-" . $params["user_info"]["username"] . "\n";
            $pushMsg .= "操作时间：" . date("Y-m-d H:i:s") . "\n\n";
            $pushMsg .= "通知来自【<a href='" . configuration("system_url") . "'>" . configuration("company_name") . "</a>】";
            $this->pushByTg($pushMsg, $config, "after_module_hard_off");
        }
    }
    public function afterModuleHardOffFailed($params)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        if ($tg_push && $config["after_module_hard_off"] == 1) {
            $msg = $params["msg"];
            $params = $params["params"];
            $this->errorFilter("after_module_hard_off_failed", "硬关机", $config, $params, $msg);
        }
    }
    public function afterModuleReboot($params)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        if ($tg_push && $config["after_module_reboot"] == 1) {
            $params = $params["params"];
            $host = \think\Db::name("host")->where("id", $params["hostid"])->find();
            $pushMsg = "产品重启成功通知\n";
            $pushMsg .= "产品名称：[" . $params["hostid"] . "]-" . $params["name"] . "\n";
            $pushMsg .= "产品IP：[<code>" . $host["dedicatedip"] . "</code>]\n\n";
            $pushMsg .= "客户姓名：[" . $params["user_info"]["id"] . "]-" . $params["user_info"]["username"] . "\n";
            $pushMsg .= "操作时间：" . date("Y-m-d H:i:s") . "\n\n";
            $pushMsg .= "通知来自【<a href='" . configuration("system_url") . "'>" . configuration("company_name") . "</a>】";
            $this->pushByTg($pushMsg, $config, "after_module_reboot");
        }
    }
    public function afterModuleRebootFailed($params)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        if ($tg_push && $config["after_module_reboot"] == 1) {
            $msg = $params["msg"];
            $params = $params["params"];
            $this->errorFilter("after_module_reboot_failed", "重启", $config, $params, $msg);
        }
    }
    public function afterModuleHardReboot($params)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        if ($tg_push && $config["after_module_hard_reboot"] == 1) {
            $params = $params["params"];
            $host = \think\Db::name("host")->where("id", $params["hostid"])->find();
            $pushMsg = "产品硬重启成功通知\n";
            $pushMsg .= "产品名称：[" . $params["hostid"] . "]-" . $params["name"] . "\n";
            $pushMsg .= "产品IP：[<code>" . $host["dedicatedip"] . "</code>]\n\n";
            $pushMsg .= "客户姓名：[" . $params["user_info"]["id"] . "]-" . $params["user_info"]["username"] . "\n";
            $pushMsg .= "操作时间：" . date("Y-m-d H:i:s") . "\n\n";
            $pushMsg .= "通知来自【<a href='" . configuration("system_url") . "'>" . configuration("company_name") . "</a>】";
            $this->pushByTg($pushMsg, $config, "after_module_hard_reboot");
        }
    }
    public function afterModuleHardRebootFailed($params)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        if ($tg_push && $config["after_module_hard_reboot"] == 1) {
            $msg = $params["msg"];
            $params = $params["params"];
            $this->errorFilter("after_module_hard_reboot_failed", "硬重启", $config, $params, $msg);
        }
    }
    public function afterModuleReinstall($params)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        if ($tg_push && $config["after_module_reinstall"] == 1) {
            $params = $params["params"];
            $host = \think\Db::name("host")->where("id", $params["hostid"])->find();
            $pushMsg = "产品重装系统成功通知\n";
            $pushMsg .= "产品名称：[" . $params["hostid"] . "]-" . $params["name"] . "\n";
            $pushMsg .= "产品IP：[<code>" . $host["dedicatedip"] . "</code>]\n\n";
            $pushMsg .= "重装系统名：" . $params["reinstall_os_name"] . "\n\n";
            $pushMsg .= "客户姓名：[" . $params["user_info"]["id"] . "]-" . $params["user_info"]["username"] . "\n";
            $pushMsg .= "操作时间：" . date("Y-m-d H:i:s") . "\n\n";
            $pushMsg .= "通知来自【<a href='" . configuration("system_url") . "'>" . configuration("company_name") . "</a>】";
            $this->pushByTg($pushMsg, $config, "after_module_reinstall");
        }
    }
    public function afterModuleReinstallFailed($params)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        if ($tg_push && $config["after_module_reinstall"] == 1) {
            $msg = $params["msg"];
            $params = $params["params"];
            $host = \think\Db::name("host")->where("id", $params["hostid"])->find();
            $pushMsg = "产品重装系统<code>失败</code>通知\n";
            $pushMsg .= "产品名称：[" . $params["hostid"] . "]-" . $params["name"] . "\n";
            $pushMsg .= "产品IP：[<code>" . $host["dedicatedip"] . "</code>]\n\n";
            $pushMsg .= "重装系统名：" . $params["reinstall_os_name"] . "\n";
            $pushMsg .= "失败原因：" . $msg . "\n\n";
            $pushMsg .= "客户姓名：[" . $params["user_info"]["id"] . "]-" . $params["user_info"]["username"] . "\n";
            $pushMsg .= "操作时间：" . date("Y-m-d H:i:s") . "\n\n";
            $pushMsg .= "通知来自【<a href='" . configuration("system_url") . "'>" . configuration("company_name") . "</a>】";
            $this->pushByTg($pushMsg, $config, "after_module_reinstall_failed");
        }
    }
    public function invoicePaid($param)
    {
        $config = \think\Db::name("plugin")->where(["name" => "WjsskPush", "module" => "addons"])->value("config");
        $config = json_decode($config, true);
        $tg_push = $config["is_tg"] ? 1 : 0;
        if ($tg_push) {
            $invoices = \think\Db::view("invoice_items II", "id,uid,type,rel_id,description,amount,payment")->view("clients C", "username", "C.id = II.uid", "LEFT")->view("invoices I", "create_time,paid_time", "I.id = II.invoice_id", "LEFT")->view("plugin P", "title as payment_name", "P.name = II.payment and module=\"gateways\"", "LEFT")->where("II.invoice_id", $param["invoiceid"])->select()->toArray();
            $curr = \think\Db::name("currencies")->find();
            if (0 < count($invoices)) {
                foreach ($invoices as $k => $invoice) {
                    $pushMsg = "";
                    $pushKey = "";
                    if ($invoice["type"] == "recharge") {
                        $pushMsg .= "充值订单付款通知";
                        $pushKey = "invoice_paid_recharge";
                    } else {
                        if ($invoice["type"] == "host") {
                            $pushMsg .= "新产品订单付款通知";
                            $pushKey = "invoice_paid_host";
                        } else {
                            if ($invoice["type"] == "renew") {
                                $pushMsg .= "续费订单付款通知";
                                $pushKey = "invoice_paid_renew";
                            }
                        }
                    }
                    if (empty($pushKey) || $config[$pushKey] != 1) {
                        continue;
                    }
                    $pushMsg .= "(" . $curr["prefix"] . " " . $invoice["amount"] . $curr["suffix"] . ")\n";
                    $pushMsg .= "账单编号：" . $param["invoiceid"] . "-" . $invoice["id"] . "\n";
                    $pushMsg .= "账单描述：[" . $invoice["description"] . "]\n";
                    $pushMsg .= "支付通道：" . (empty($invoice["payment"]) ? "余额支付" : $invoice["payment_name"]) . "\n";
                    $pushMsg .= "实际支付：" . $curr["prefix"] . " " . $invoice["amount"] . $curr["suffix"] . "\n";
                    $pushMsg .= "客户姓名：" . $invoice["username"] . "\n";
                    $pushMsg .= "订单创建时间：" . date("Y-m-d H:i:s", $invoice["create_time"]) . "\n";
                    $pushMsg .= "订单支付时间：" . date("Y-m-d H:i:s", $invoice["paid_time"]) . "\n\n";
                    $pushMsg .= "通知来自【<a href='" . configuration("system_url") . "'>" . configuration("company_name") . "</a>】";
                    $this->pushByTg($pushMsg, $config, "invoice_paid");
                }
            }
        }
    }
    public function pushByTg($content, $config, $type)
    {
        try {
            $bot_token = $config["tg_bot_token"];
            $chat_id = explode(",", $config["tg_user_id"]);
            $tgApiUrl = $config["tg_proxy"] . "https://api.telegram.org/bot" . $bot_token . "/sendMessage";
            foreach ($chat_id as $k => $c_id) {
                $post_fields = ["chat_id" => $c_id, "text" => $content, "parse_mode" => "HTML"];
                $ch = curl_init();
                curl_setopt($ch, CURLOPT_URL, $tgApiUrl);
                curl_setopt($ch, CURLOPT_POST, true);
                curl_setopt($ch, CURLOPT_POSTFIELDS, $post_fields);
                curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
                $response = curl_exec($ch);
                curl_close($ch);
                $res = json_decode($response, true);
                if ($res["ok"]) {
                    $this->iLog(json_encode($post_fields, JSON_UNESCAPED_UNICODE), $type);
                }
            }
        } catch (\Exception $e) {
            $this->iLog(json_encode($post_fields, JSON_UNESCAPED_UNICODE), $type, "ERROR");
        }
    }
    public function iLog($content, $type, $flag = "SUCCESS")
    {
        $fileName = dirname(__DIR__) . "/wjssk_push/log/" . date("Ymd");
        $log = ["date_time" => date("Y-m-d H:i:s"), "content" => $content, "result" => $flag, "type" => $type];
        \think\Db::startTrans();
        try {
            \think\Db::name("wjssk_push_log")->insert($log);
            \think\Db::commit();
        } catch (\Exception $e) {
            \think\Db::rollback();
            file_put_contents($fileName, date("Y-m-d") . "----\n存入日志失败：" . $e->getMessage() . "\n" . json_encode($log) . "\n" . "\n", FILE_APPEND);
        }
    }
}

?>