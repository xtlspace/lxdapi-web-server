<?php
namespace addons\wjssk_push\controller;

class WjsskPushController extends \app\admin\controller\PluginAdminBaseController
{
    private $_config = [];
    private $_info = [];
    private $lang;
    public function initialize()
    {
        parent::initialize();
        $this->_config = $this->getPlugin()->getConfig();
        $this->_info = $this->getPlugin()->info;
        $lang = request()->languagesys;
        if (empty($lang)) {
            $lang = configuration("language") ? configuration("language") : config("default_lang");
        }
        $this->lang = $lang;
    }
    public function set()
    {
        $this->assign("config", $this->_config);
        return $this->fetch("/set");
    }
    public function saveSet()
    {
        \think\Db::startTrans();
        try {
            $param = request()->post();
            $config = \think\Db::name("plugin")->where("name", $this->_info["name"])->value("config");
            if (empty($config)) {
                $config = [];
            }
            $config = json_decode($config, true);
            $result = array_replace($config, $param);
            \think\Db::name("plugin")->where("name", $this->_info["name"])->update(["config" => json_encode($result)]);
            \think\Db::commit();
        } catch (\Exception $e) {
            \think\Db::rollback();
            $this->error("配置保存失败！Error：" . $e->getMessage());
        }
        $this->success("配置保存成功！");
    }
    public function log()
    {
        return $this->fetch("/log");
    }
    public function showLog()
    {
        $search = request()->post("search");
        $page = request()->post("page");
        $limit = request()->post("limit");
        $start = $search["start"];
        $end = $search["end"];
        $where = "(1=1)";
        if (!empty($start) && !empty($end)) {
            if ($start == $end) {
                $where = "date_time like '" . $end . "%'";
            } else {
                $where = "Date(date_time) BETWEEN '" . $start . "' and '" . $end . "'";
            }
        }
        try {
            $total = \think\Db::name("wjssk_push_log")->where($where)->count();
            $log = \think\Db::name("wjssk_push_log")->where($where)->page($page, $limit)->order("date_time desc")->select();
        } catch (\Exception $e) {
            $this->error("获取日志失败！Error:" . $e->getMessage());
        }
        $this->success("success", "", ["data" => $log, "page" => $page, "limit" => $limit, "total" => $total]);
    }
}

?>