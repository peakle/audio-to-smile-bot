<?php

declare(strict_types=1);

namespace App\Service;

use Predis\Client as Redis;

class QueueService
{
    /**
     * @var Redis
     */
    protected $queueClient;

    /**
     * @param string $redisHost
     * @param string $redisPort
     */
    public function __construct(string $redisHost, string $redisPort)
    {
        $this->queueClient = new Redis([
            'scheme' => 'tcp',
            'host' => $redisHost,
            'port' => $redisPort
        ]);
    }

    /**
     * @param string $queueName
     * @param string $data
     */
    public function put(string $queueName, string $data): void
    {
        $this->queueClient->set($queueName, $data);
    }

    /**
     * @param string $queueName
     * @return string
     */
    public function take(string $queueName): string
    {
        return $this->queueClient->get($queueName);
    }
}
