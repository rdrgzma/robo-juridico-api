<?php

use Illuminate\Support\Str;

return [
    'default' => env('DB_CONNECTION', 'sqlite'),

    'connections' => [
        'sqlite' => [
            'driver' => 'sqlite',
            'database' => (env('DB_DATABASE') === ':memory:' || (env('DB_DATABASE') && (str_starts_with(env('DB_DATABASE'), '/') || str_starts_with(env('DB_DATABASE'), '\\') || preg_match('/^[A-Za-z]:/', env('DB_DATABASE')))))
                ? env('DB_DATABASE')
                : base_path(env('DB_DATABASE', 'database/database.sqlite')),
            'foreign_key_constraints' => env('DB_FOREIGN_KEYS', true),
            'busy_timeout' => 15000,
            'journal_mode' => 'WAL',
            'synchronous' => 'NORMAL',
        ],
    ],

    'migrations' => [
        'table' => 'migrations',
        'update_date_on_publish' => true,
    ],
];
