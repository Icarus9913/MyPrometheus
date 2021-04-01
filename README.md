## prometheus

* 1.官方node_exporter写法
  

* 2.OpenCensus写法
  > - OpenCensus也有tracing那个功能,同时关注一下OpenTracing  
  > - 代码层面先暴露一个http的路由,然后通过暴露出来,然后,在prometheus的yaml还是toml配置文件里写入这个暴露的端口路由,prometheus就可以监听数据了.  
  > - 另外可以通过promQL写一些规则,根据数据触发报警. 通过AlterManager去报警
  > - 还有一个push gateway能够主动的在代码里,将数据push给prometheus,而不是通过promtheus server去向端口路由拿数据


* Prometheus:
  >  - Counter: 表示一个累计度量，只增不减，重启后恢复为0。适用于访问次数统计，异常次数统计等场景。 
  >  - Gauge: 表示可变化的度量值，适用于CPU,内存使用率等
  >  - Histogram: 对指标的范围性（区间）统计。比如内存在0%-30%，30%-70%之间的采样次数。  
         Histogram包含三个指标：  
         * <basename>：度量值名称  
         * <basename>_count： 样本发生总次数  
         * <basename>_sum：样本发生次数中值的总和  
         * <basename>_bucket{le="+Inf"}： 每个区间的样本数    其中inf表示无穷值 
  >  - Summary: 和histogram类似，提供次数和总和，同时提供每个滑动窗口中的分位数。

* OpenCensus:
  > - Tags associate with metrics;;;   organize metrics into a View;;;  Export views to a backend(Prometheus)  
  > - histogram是柱状图，在Prometheus系统中的查询语言中，有三种作用：1. 对每个采样点进行统计（并不是一段时间的统计），打到各个桶(bucket)中; 2.对每个采样点值累计和(sum); 3.对采样点的次数累计和(count)
  > - Summary 与 Histogram 类似，会对观测数据进行取样，得到数据的个数和总和。此外，还会取一个滑动窗口，计算窗口内样本数据的{百分位数}。 更耗性能.   
      如果需要聚合,则选择histogram  
      * [basename]_count: 数据的个数，类型为 counter  
      * [basename]_sum: 数据的加和，类型为 counter  
      * [basename]{quantile=0.5}: 滑动窗口内 50% 分位数值  
      * [basename]{quantile=0.9}: 滑动窗口内 90% 分位数值  
      * [basename]{quantile=0.99}: 滑动窗口内 99% 分位数值  

* PromQL: up指标可以获取到当前所有运行的Exporter实例以及其状态  

> 1.直接使用监控指标名称查询时，可以查询该指标下的所有时间序列:  
http_requests_total  等同于	http_requests_total{}  
返回:  
http_requests_total{code="200",handler="alerts",instance="localhost:9090",job="prometheus",method="get"}=(20889@1518096812.326)  
http_requests_total{code="200",handler="graph",instance="localhost:9090",job="prometheus",method="get"}=(21287@1518096812.326)  

> 2.匹配与排除: =和!=  
查询所有http_requests_total时间序列中满足标签instance为localhost:9090的时间序列:  
http_requests_total{instance="localhost:9090"}		反之排除:http_requests_total{instance!="localhost:9090"}  

> 3.正则: 匹配是label='~'regx  反之排除使用 label!'~'regx  
http_requests_total{environment='~'"staging|testing|development",method!="GET"}  

> 4.范围查询:  
-时间范围查询:  
近5分钟内所有数据:http_requests_total{}[5m]  
-时间位移操作:  
例5分钟之前的瞬时样本数据: http_request_total{} offset 5m  
昨天一天的区间内样本数据:  http_request_total{}[1d] offset 1d  

> 5.聚合操作:  
-总量:	sum(http_request_total)  
-平均(按照mode计算主机CPU的平均使用时间):	avg(node_cpu) by (mode)  
-按照主机查询各个主机的CPU使用率:	sum(sum(irate(node_cpu{mode!='idle'}[5m])) / sum(irate(node_cpu[5m]))) by (instance)  

> 6.数学运算:  
-(加法+): node_disk_bytes_written + node_disk_bytes_read  
-(减法-):  
-(乘法*):  
-(除法/): node_memory_free_bytes_total / (1024 * 1024)  
-(求余%):  
-(幂运算^):  

> 7.布尔运算符:  
-(相等==)  
-(不想等!=)  
-(大于>)  
-(小于<)  
-(大于等于>=)  
-(小于等于<=)  

> 8.布尔修饰符:  
-例如:  2 == bool 2 # 结果为1  
-例如判断当前模块的HTTP请求量是否>=1000，如果大于等于1000则返回1（true）否则返回0（false）:  
http_requests_total > bool 1000  
返回:  
http_requests_total{code="200",handler="query",instance="localhost:9090",job="prometheus",method="get"}  1  
http_requests_total{code="200",handler="query_range",instance="localhost:9090",job="prometheus",method="get"}  0  

> 9.集合运算符:  
-(并且and) 	A and B 	就是A交B  
-(或者or)  	A or B  	就是A并B  
-(排除unless)  A unless B  就是A-(A交B)  

>10.操作符优先级(高到低):  
 > ^  
 > '*'  /  %      
 > '+'  -  
 > ==  !=  <=  <  >=  >    
 > and  unless   
 > or    

> 11.匹配模式  
在操作符两边表达式标签不一致的情况下，可以使用on(label list)或者ignoring(label list）来修改便签的匹配行为。使用ignoreing可以在匹配时忽略某些便签。而on则用于将匹配行为限定在某些便签之内。    
--一对一(one-to-one)    
例如:method_code:http_errors:rate5m{code="500"} / ignoring(code) method:http_requests:rate5m    
该表达式会返回在过去5分钟内，HTTP请求状态码为500的在所有请求中的比例。如果没有使用ignoring(code)，操作符两边表达式返回的瞬时向量中将找不到任何一个标签完全相同的匹配项。    
返回:  
{method="get"}  0.04            //  24 / 600  
{method="post"} 0.05            //   6 / 120  
--多对一(many-to-one)或一对多(one-to-many)  
多对一和一对多两种匹配模式指的是“一”侧的每一个向量元素可以与"多"侧的多个元素匹配的情况。在这种情况下，必须使用group修饰符：group_left或者group_right来确定哪一个向量具有更高的基数（充当“多”的角色    
例如:method_code:http_errors:rate5m / ignoring(code) group_left method:http_requests:rate5m    
该表达式中，左向量method_code:http_errors:rate5m包含两个标签method和code。而右向量method:http_requests:rate5m中只包含一个标签method，因此匹配时需要使用ignoring限定匹配的标签为code。 在限定匹配标签后，右向量中的元素可能匹配到多个左向量中的元素 因此该表达式的匹配模式为多对一，需要使用group修饰符group_left指定左向量具有更好的基数。    

> 12.聚合操作  
-(sum求和)  
-(min最小值)  
-(max最大值)  
-(avg平均值)  
-(stddev标准差)  
-(stdvar标准方差)  
-(count计数)  
-(count_values对value进行计数)  
-(bottomk后n条时序)  
-(topk前n条时序)  
-(quantile分位数)  
聚合操作语法:<aggr-op>([parameter,] <vector expression>) [without|by (<label list>)]  
其中只有count_values, quantile, topk, bottomk支持参数(parameter)。  
without用于从计算结果中移除列举的标签，而保留其它标签。  
by结果向量中只保留列出的标签，其余标签则移除。  
例如:  
sum(http_requests_total) without (instance)  等价于  sum(http_requests_total) by (code,handler,job,method)  

> 13.内置函数  
-increase(v range-vector) 增长量   
例如: increase(node_cpu[2m]) / 120       
解释: 这里通过node_cpu[2m]获取时间序列最近两分钟的所有样本，increase计算出最近两分钟的增长量，最后除以时间120秒得到node_cpu样本在最近两分钟的平均增长率. <br>    
-rate(v range-vector)  平均增长速率  
例如: rate(node_cpu[2m]) 和上面的例子得到的结果一样 <br>  
-irate(v range-vector)  瞬时增长率  
例如: irate(node_cpu[2m])   
irate取的是在指定时间范围内的最近两个数据点来算速率，而rate会取指定时间范围内所有数据点，算出一组速率，然后取平均值作为结果。irate适合快速变化的计数器（counter），而rate适合缓慢变化的计数器（counter）。 <br>  
-predict_linear(v range-vector, t scalar)  预测时间序列v在t秒后的值    
例如: predict_linear(node_filesystem_free{job="node"}[2h], 4 * 3600) < 0    
解释: 基于2小时的样本数据，来预测主机可用磁盘空间的是否在4个小时候被占满   <br>  
-histogram_quantile(φ float, b instant-vector)  Histogram计算分位数  
例如: histogram_quantile(0.5, http_request_duration_seconds_bucket)  
解释:计算中位数   <br>  
-label_replace(v instant-vector, dst_label string, replacement string, src_label string, regex string)    
解释:   <br>
-abs(v instant-vector)    
解释:返回输入向量的所有样本的绝对值   <br>  
-absent(v instant-vector)     
解释:输入有值时,返回空向量.  输入样本无值时,返回1   <br>  
-absent_over_time(v range-vector)   <br>  
-ceil(v instant-vector)   
解释:是一个向上舍入为最接近的整数。  <br>  
-changes(v range-vector)    
解释:输入一个范围向量， 返回这个范围向量内每个样本数据值变化的次数。 <br>  
-clamp_max(v instant-vector, max scalar)  
解释:输入一个瞬时向量和最大值，样本数据值若大于max，则改为max，否则不变   <br>  
-clamp_min(v instant-vector)  
解释:输入一个瞬时向量和最大值，样本数据值小于min，则改为min。否则不变   <br>  
-day_of_month(v=vector(time()) instant-vector)  
解释:返回被给定UTC时间所在月的第几天。返回值范围：1-31。 <br>  
-day_of_week(v=vector(time()) instant-vector)  
解释:返回被给定UTC时间所在周的第几天。返回值范围：0-6. 0表示星期天   <br>  
-days_in_month(v=vector(time()) instant-vector)  
解释:返回当月一共有多少天。返回值范围：28-31   <br>  
-delta(v range-vector)  
解释:计算一个范围向量v的第一个元素和最后一个元素之间的差值  
例如:delta(cpu_temp_celsius{host="zeus"}[2h])	返回过去两小时的CPU温度差   <br>  
-deriv(v range-vector)  
解释:计算一个范围向量v中各个时间序列二阶导数   <br>  
-exp(v instant-vector)  
解释:输入一个瞬时向量, 返回各个样本值的e指数值，即为e^N次方   <br>  
-floor(v instant-vector)  
解释:与ceil()函数相反。 向下取整 4.3 为 4   <br>  
-histogram_quantile(φ scalar, b instant-vector)  
解释:计算b向量的φ-直方图 (0 ≤ φ ≤ 1)   <br>
-holt_winters(v range-vector, sf scalar, tf scalar)  
解释:基于范围向量v，生成事件序列数据平滑值。平滑因子sf越低, 对老数据越重要。趋势因子tf越高，越多的数据趋势应该被重视。0< sf, tf <=1。 holt_winters仅用于gauges   <br>  
-hour(v=vector(time()) instant-vector)  
解释:返回被给定UTC时间的当前第几个小时，时间范围：0-23。   <br>  
-idelta(v range-vector)  
解释:输入一个范围向量，返回key: value = 度量指标： 每最后两个样本值差值。  <br>    
-label_join(v instant-vector, dst_label string, separator string, src_label_1 string, src_label_2 string, ...)   <br>  
-label_replace(v instant-vector, dst_label string, replacement string, src_label string, regex string)   <br>  
-ln(v instant-vector)   
解释:计算瞬时向量v中所有样本数据的自然对数   <br>    
-log2(v instant-vector)  
解释:计算瞬时向量v中所有样本数据的二进制对数。   <br>  
-log10(v instant-vector)  
解释:计算瞬时向量v中所有样本数据的10进制对数。相当于ln()    <br>  
-minute(v=vector(time()) instant-vector)  
解释:返回给定UTC时间当前小时的第多少分钟。结果范围：0-59。   <br>  
-month(v=vector(time()) instant-vector)  
解释:返回给定UTC时间当前属于第几个月，结果范围：0-12。   <br>  
-predict_linear(v range-vector, t scalar)  
解释:预测函数,输入范围向量和从现在起t秒; 输出不带有度量指标,只有标签列表的结果值  
例如:predict_linear(http_requests_total{code="200",instance="120.77.65.193:9090",job="prometheus",method="get"}[5m], 5)   <br>  
-resets(v range-vector)  
解释:输入一个范围向量;输出一个具有标签列表[在这个范围向量中每个度量指标被重置的次数]. 在两个连续样本数据值下降,也可以理解为counter被重置  
例如:resets(http_requests_total[5m])   <br>  
-round(v instant-vector, to_nearest=1 scalar)  
解释:输入瞬时向量; 输出指定整数级的四舍五入值,如果不指定,则是1以内的四舍五入.   <br>  
-scalar(v instant-vector)  
解释:输入瞬时向量; 输出：key: value = "scalar", 样本值[如果度量指标样本数量大于1或者等于0, 则样本值为NaN, 否则，样本值本身]   <br>  
-sort(v instant-vector)  
解释:输入瞬时向量; 输出：key: value = 度量指标：样本值[升序排列]   <br>  
-sort_desc(v instant-vector)  
解释:输入瞬时向量; 输出：key: value = 度量指标：样本值[降序排列]   <br>  
-sqrt(v instant-vector)  
解释:输入瞬时向量; 输出：key: value = 度量指标：样本值的平方根   <br>  
-time()  
解释:返回从1970-01-01到现在的秒数，注意：它不是直接返回当前时间，而是时间戳   <br>  
-timestamp(v instant-vector)  
解释:返回给定向量的每个样本的时间戳,为1970年1月1日UTC以来的秒数   <br>  
-vector(s scalar)  
解释:returns the scalar s as a vector with no labels.   <br>  
-year(v=vector(time()) instant-vector)  
解释:返回年份   <br>  

>14.HTTP响应数据类型:  
-vector瞬时向量  
-matrix区间向量  
-scalar标量  
-string字符串    

>15.HTTP区间数据查询  
-QUERY_RANGE  
-query=: PromQL表达式。  
-start=: 起始时间。  
-end=: 结束时间。  
-step=: 查询步长。  
-timeout=: 超时设置。可选参数，默认情况下使用-query,timeout的全局设置。                  
例如:$ curl 'http://localhost:9090/api/v1/query_range?query=up&start=2015-07-01T20:10:30.781Z&end=2015-07-01T20:11:00.781Z&step=15s'  

```shell
在yaml配置文件中增加label标签:
scrape_configs:
  - job_name: 'node_exporter'
    static_configs:
    - targets: ['192.168.2.101:9100']
      labels:       
        instance: watch_me_ok?
```
