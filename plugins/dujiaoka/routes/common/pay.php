<?php
/**
 * The file was created by Assimon.
 *
 * @author    assimon<ashang@utf8.hk>
 * @copyright assimon<ashang@utf8.hk>
 * @link      http://utf8.hk/
 */
use Illuminate\Support\Facades\Route;

Route::get('pay-gateway/{handle}/{payway}/{orderSN}', 'PayController@redirectGateway');

// 支付相关
Route::group(['prefix' => 'pay', 'namespace' => 'Pay', 'middleware' => ['dujiaoka.pay_gate_way']], function () {

    /**
     这里是原来的代码
     */

    // USDTMore
    Route::get('usdtmore/{payway}/{orderSN}', 'USDTMoreController@gateway');
    Route::post('usdtmore/notify_url', 'USDTMoreController@notifyUrl');
    Route::get('usdtmore/return_url', 'USDTMoreController@returnUrl')->name('usdtmore-return');

});
