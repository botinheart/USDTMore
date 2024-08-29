<?php
declare (strict_types=1);

return [
    'version' => '1.0.0',
    'name' => 'USDTMore',
    'author' => '森海北屿',
    'website' => 'https://ovsea.net',
    'description' => 'USDTMore，一款好用的多链路个人USDT收款网关',
    'options' => [
        'TRON'  => 'TRC20',
        'BSC'   => 'BEP20',
        'POLY'  => 'Polygon',
        'OP'    => 'Optimism'
    ],
    'callback' => [
        \App\Consts\Pay::IS_SIGN => true,
        \App\Consts\Pay::IS_STATUS => true,
        \App\Consts\Pay::FIELD_STATUS_KEY => 'status',
        \App\Consts\Pay::FIELD_STATUS_VALUE => 2,
        \App\Consts\Pay::FIELD_ORDER_KEY => 'order_id',
        \App\Consts\Pay::FIELD_AMOUNT_KEY => 'amount',
        \App\Consts\Pay::FIELD_RESPONSE => 'ok'
    ]
];