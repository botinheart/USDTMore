<?php
declare(strict_types=1);

namespace App\Pay\USDTMore\Impl;

use App\Entity\PayEntity;
use App\Pay\Base;
use GuzzleHttp\Exception\GuzzleException;
use Kernel\Exception\JSONException;

/**
 * Class Pay
 * @package App\Pay\Kvmpay\Impl
 */
class Pay extends Base implements \App\Pay\Pay
{

    /**
     * @return PayEntity
     * @throws JSONException
     */
    public function trade(): PayEntity
    {

        if (!$this->config['url']) {
            throw new JSONException("请配置网关地址");
        }

        if (!$this->config['key']) {
            throw new JSONException("请配置密钥");
        }

        $param = [
            'code' => $this->code,
            'order_id' => $this->tradeNo,
            'amount' => $this->amount,
            'notify_url' => $this->callbackUrl,
            'redirect_url' => $this->returnUrl
        ];

        $param['signature'] = Signature::generateSignature($param, $this->config['key']);

        try {
            $request = $this->http()->post(trim($this->config['url'], "/") . '/api/v1/order/create-transaction', [
                "json" => $param
            ]);
        } catch (GuzzleException $e) {
            throw new JSONException("网关(".trim($this->config['url'], "/") . '/api/v1/order/create-transaction'.")连接失败，下单未成功");
        }

        $contents = $request->getBody()->getContents();
        $json = (array)json_decode((string)$contents, true);

        if ($json['status_code'] != 200) {
            throw new JSONException((string)$json['message']);
        }

        $payEntity = new PayEntity();
        $payEntity->setType(self::TYPE_REDIRECT);
        $payEntity->setUrl($json['data']['payment_url']);
        return $payEntity;
    }
}