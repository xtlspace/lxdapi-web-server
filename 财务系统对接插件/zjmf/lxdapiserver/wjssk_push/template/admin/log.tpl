<link rel="stylesheet" href="https://unpkg.com/element-ui/lib/theme-chalk/index.css">
<style>
    .wjssk-app .el-date-editor .el-range-separator{
        box-sizing: unset;
    }
</style>
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
                    <div class="wjssk-app" id="wjssk_app">
                        <div class="block">
                            <el-date-picker
                                    v-model="searchVal"
                                    type="daterange"
                                    range-separator="至"
                                    start-placeholder="开始日期"
                                    end-placeholder="结束日期"
                                    value-format="yyyy-MM-dd"
                                    size="small">
                            </el-date-picker>
                            <el-button type="primary" size="small" @click="doSearch">筛选</el-button>
                        </div>
                        <el-divider></el-divider>
                        <el-table
                                v-loading="tableLoading"
                                :data="logData"
                                border
                                style="width: 100%"
                                size="small"
                                :header-cell-style="{background:'#eef1f6',color:'#606266'}">
                            <el-table-column
                                    fixed
                                    prop="id"
                                    label="#"
                                    width="100">
                            </el-table-column>
                            <el-table-column
                                    fixed
                                    prop="date_time"
                                    label="日志时间"
                                    width="200">
                            </el-table-column>
                            <el-table-column
                                    prop="result"
                                    label="日志状态"
                                    width="120">
                            </el-table-column>
                            <el-table-column
                                    prop="type"
                                    label="日志类型"
                                    width="200">
                            </el-table-column>
                            <el-table-column
                                    prop="content"
                                    label="日志内容">
                            </el-table-column>
                        </el-table>
                        <el-pagination
                                background
                                @size-change="handleSizeChange"
                                @current-change="handleCurrentChange"
                                :current-page.sync="tableData.page"
                                :page-sizes="[20, 30, 40, 50]"
                                :page-size.sync="tableData.limit"
                                layout="sizes, prev, pager, next"
                                :total="tableData.total">
                        </el-pagination>
                    </div>
                </div>
            </div>
        </div>
    </div>
</section>
<script src="https://cdn.jsdelivr.net/npm/vue@2.7.14/dist/vue.js"></script>
<script src="https://unpkg.com/element-ui/lib/index.js"></script>
<script>
    const app = new Vue({
        el: '#wjssk_app',
        data:{
            searchVal:'',
            logData:[],
            tableData:{
                total:0,
                page:1,
                limit:20
            },
            tableLoading:true
        },
        created(){
            this.getLogData();
        },
        methods:{
            doSearch(){
                this.tableData.total = 0;
                this.tableData.page = 1;
            },
            getLogData(){
                let start = '',
                    end = '';
                if(this.searchVal && this.searchVal.length === 2){
                    start = this.searchVal[0];
                    end = this.searchVal[1];
                }
                let page = this.tableData.page;
                let limit = this.tableData.limit;
                let _this = this;
                _this.tableLoading = true;
                $.post('{:shd_addon_url('WjsskPush://WjsskPush/showLog')}',{
                    search:{start,end},
                    page,limit
                },function (res) {
                    if(res.code){
                        _this.logData = res.data.data;
                        _this.tableData.total = res.data.total;
                        _this.tableLoading = false;
                    }else{
                        _this.$message.error(res.msg);
                    }
                },'json')
            },
            handleSizeChange(){
                // 调整显示条数
                console.log(this.tableData.limit);
            },
            handleCurrentChange(){
                // 页数
                console.log(this.tableData.page);
            }
        }
    })
</script>
