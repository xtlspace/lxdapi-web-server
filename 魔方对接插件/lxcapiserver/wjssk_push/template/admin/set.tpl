<section class="admin-main">
    <div class="container-fluid">
        <div class="page-container">
            <div class="card">
                <div class="card-body">
                    <div class="card-title row"> <div style="padding:0 15px;">{:lang('title')}</div>
                        <!-- 页面菜单 开始 如果需要,只需将此段代码放入tpl文件-->
                        <div class="col-lg-8 col-md-12 col-sm-12">
                            {foreach $PluginsAdminMenu as $v}
                                {if $v['custom']}
                                    <span  class="ml-2"><a  class="h5" href="{$v.url}" target="_blank">{$v.name}</a></span>
                                {else/}
                                    <span  class="ml-2"> <a  class="h5" href="{$v.url}">{$v.name}</a></span>
                                {/if}
                            {/foreach}
                        </div>
                        <!-- 页面菜单 结束 -->
                    </div>
                    <div class="help-block">
                        {:lang('help')}QQ：484449540，TG：kid_jok
                    </div>

                    <form action="{:shd_addon_url('WjsskPush://WjsskPush/saveSet')}" id="form_saveSet">
                        <div class="card-body px-5 mx-auto w-75">
                            <div class="form-group row">
                                <label class="col-sm-2 col-form-label text-right">TG机器人消息通知</label>
                                <div class="col-sm-10 custom-control custom-switch">
                                    <input type="hidden" name="is_tg" value="0">
                                    <input type="checkbox" class="custom-control-input" id="is_tg" name="is_tg" value="1" {$config.is_tg?'checked':''}>
                                    <label class="custom-control-label" for="is_tg"></label>
                                </div>
                            </div>
                            <div class="form-group row">
                                <label for="tg_bot_token" class="col-sm-2 col-form-label text-right">TG机器人Token</label>
                                <div class="col-sm-10 pl-0">
                                    <input type="text" class="form-control" id="tg_bot_token" name="tg_bot_token" value="{$config.tg_bot_token}" placeholder="输入TG机器人Token" autocomplete="off">
                                </div>
                            </div>
                            <div class="form-group row">
                                <label for="tg_user_id" class="col-sm-2 col-form-label text-right">TG用户Id</label>
                                <div class="col-sm-10 pl-0">
                                    <input type="text" class="form-control" id="tg_user_id" name="tg_user_id" value="{$config.tg_user_id}" placeholder="输入TG用户Id" autocomplete="off">
                                </div>
                            </div>
                            <div class="form-group row">
                                <label for="tg_user_id" class="col-sm-2 col-form-label text-right">TG代理url</label>
                                <div class="col-sm-10 pl-0">
                                    <input type="text" class="form-control" id="tg_proxy" name="tg_proxy" value="{$config.tg_proxy}" placeholder="输入TG代理url" autocomplete="off">
                                </div>
                            </div>
                            <hr>
                            <div class="form-group row">
                                <label for="tg_user_id" class="col-sm-2 col-form-label text-right">通知范围</label>
                                <div class="col-sm-10 pl-0" style="display: flex;flex-wrap: wrap;align-items: center;">
                                    <div class="form-check form-check-inline">
                                        <input type="hidden" name="invoice_paid_host" value="0">
                                        <input class="form-check-input" type="checkbox" name="invoice_paid_host" id="invoice_paid_host" value="1" {eq name="$config.invoice_paid_host" value="1"}checked{/eq}>
                                        <label class="form-check-label" for="invoice_paid_host">新产品订单付款通知</label>
                                    </div>
                                    <div class="form-check form-check-inline">
                                        <input type="hidden" name="invoice_paid_renew" value="0">
                                        <input class="form-check-input" type="checkbox" name="invoice_paid_renew" id="invoice_paid_renew" value="1" {eq name="$config.invoice_paid_renew" value="1"}checked{/eq}>
                                        <label class="form-check-label" for="invoice_paid_renew">续费订单付款通知</label>
                                    </div>
                                    <div class="form-check form-check-inline">
                                        <input type="hidden" name="invoice_paid_recharge" value="0">
                                        <input class="form-check-input" type="checkbox" name="invoice_paid_recharge" id="invoice_paid_recharge" value="1" {eq name="$config.invoice_paid_recharge" value="1"}checked{/eq}>
                                        <label class="form-check-label" for="invoice_paid_recharge">充值订单付款通知</label>
                                    </div>
                                    <div class="form-check form-check-inline">
                                        <input type="hidden" name="after_module_create" value="0">
                                        <input class="form-check-input" type="checkbox" name="after_module_create" id="after_module_create" value="1" {eq name="$config.after_module_create" value="1"}checked{/eq}>
                                        <label class="form-check-label" for="after_module_create">产品开通通知</label>
                                    </div>
                                    <div class="form-check form-check-inline">
                                        <input type="hidden" name="after_module_on" value="0">
                                        <input class="form-check-input" type="checkbox" name="after_module_on" id="after_module_on" value="1" {eq name="$config.after_module_on" value="1"}checked{/eq}>
                                        <label class="form-check-label" for="after_module_on">产品开机操作通知</label>
                                    </div>
                                    <div class="form-check form-check-inline">
                                        <input type="hidden" name="after_module_off" value="0">
                                        <input class="form-check-input" type="checkbox" name="after_module_off" id="after_module_off" value="1" {eq name="$config.after_module_off" value="1"}checked{/eq}>
                                        <label class="form-check-label" for="after_module_off">产品关机操作通知</label>
                                    </div>
                                    <div class="form-check form-check-inline">
                                        <input type="hidden" name="after_module_reboot" value="0">
                                        <input class="form-check-input" type="checkbox" name="after_module_reboot" id="after_module_reboot" value="1" {eq name="$config.after_module_reboot" value="1"}checked{/eq}>
                                        <label class="form-check-label" for="after_module_reboot">产品重启操作通知</label>
                                    </div>
                                    <div class="form-check form-check-inline">
                                        <input type="hidden" name="after_module_hard_off" value="0">
                                        <input class="form-check-input" type="checkbox" name="after_module_hard_off" id="after_module_hard_off" value="1" {eq name="$config.after_module_hard_off" value="1"}checked{/eq}>
                                        <label class="form-check-label" for="after_module_hard_off">产品硬关机操作通知</label>
                                    </div>
                                    <div class="form-check form-check-inline">
                                        <input type="hidden" name="after_module_hard_reboot" value="0">
                                        <input class="form-check-input" type="checkbox" name="after_module_hard_reboot" id="after_module_hard_reboot" value="1" {eq name="$config.after_module_hard_reboot" value="1"}checked{/eq}>
                                        <label class="form-check-label" for="after_module_hard_reboot">产品硬重启操作通知</label>
                                    </div>
                                    <div class="form-check form-check-inline">
                                        <input type="hidden" name="after_module_reinstall" value="0">
                                        <input class="form-check-input" type="checkbox" name="after_module_reinstall" id="after_module_reinstall" value="1" {eq name="$config.after_module_reinstall" value="1"}checked{/eq}>
                                        <label class="form-check-label" for="after_module_reinstall">产品重装系统操作通知</label>
                                    </div>
                                    <div class="form-check form-check-inline">
                                        <input type="hidden" name="after_module_crack_password" value="0">
                                        <input class="form-check-input" type="checkbox" name="after_module_crack_password" id="after_module_crack_password" value="1" {eq name="$config.after_module_crack_password" value="1"}checked{/eq}>
                                        <label class="form-check-label" for="after_module_crack_password">产品重置密码操作通知</label>
                                    </div>
                                    <div class="form-check form-check-inline">
                                        <input type="hidden" name="ticket_open" value="0">
                                        <input class="form-check-input" type="checkbox" name="ticket_open" id="ticket_open" value="1" {eq name="$config.ticket_open" value="1"}checked{/eq}>
                                        <label class="form-check-label" for="ticket_open">用户创建工单通知</label>
                                    </div>
                                    <div class="form-check form-check-inline">
                                        <input type="hidden" name="ticket_user_reply" value="0">
                                        <input class="form-check-input" type="checkbox" name="ticket_user_reply" id="ticket_user_reply" value="1" {eq name="$config.ticket_user_reply" value="1"}checked{/eq}>
                                        <label class="form-check-label" for="ticket_user_reply">用户回复工单通知</label>
                                    </div>
                                </div>
                            </div>
                            <hr>
                            <div class="col-sm-12 pr-0">
                                <div class="form-group mb-0 text-right">
                                    <button type="submit" class="btn btn-primary w-md submitBtn">
                                        <span class="spinner-border spinner-border-sm" hidden role="status" aria-hidden="true"></span>
                                        保存
                                    </button>
                                </div>
                            </div>
                        </div>
                    </form>
                </div>
            </div>
        </div>
    </div>
</section>
<script>
 $('#form_saveSet').submit(function (e){
     e.preventDefault();
     $('.submitBtn').attr('disabled',true);
     $('.submitBtn span').removeAttr('hidden');
     $.post($(this).attr('action'),$(this).serialize(),function(res){
         if(res.code){
             toastr.success(res.msg);
         }else{
             toastr.error(res.msg);
         }
         $('.submitBtn').removeAttr('disabled');
         $('.submitBtn span').attr('hidden',true);
     },'json')
 });
</script>