## 使用方法

1.将`routes`目录里面的代码增加到独角数卡的route上

    // USDTMore
    Route::get('usdtmore/{payway}/{orderSN}', 'USDTMoreController@gateway');
    Route::post('usdtmore/notify_url', 'USDTMoreController@notifyUrl');
    Route::get('usdtmore/return_url', 'USDTMoreController@returnUrl')->name('usdtmore-return');

2.将app/Http/Controllers/Pay/USDTMoreController.php复制到独角数卡的plugins/dujiaoka/app/Http/Controllers/Pay/目录下

2.在独角数卡后台添加您需要的支付方式。      

| 支付选项     | 商户id | 商户key | 商户密钥 | 支付标识               | 备注                                                                                        |     
|:---------| :----- | :----- | :----- |--------------------|:------------------------------------------------------------------------------------------|       
| USDTMore | api接口认证token	 | 空 | epusdt收银台地址+/api/v1/order/create-transaction| TRON\|POLY\OP\|BSC | 如果独角数卡和epusdt在同一服务器则填写`127.0.0.1`不要填域名，例如`http://127.0.0.1:8000/api/v1/order/create-transaction` |