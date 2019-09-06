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

    public function __construct(string $redisHost, string $redisPort)
    {
        $this->queueClient = new Redis([
            'scheme' => 'tcp',
            'host' => $redisHost,
            'port' => $redisPort
        ]);
    }

    public function put(string $queueName, string $data): void
    {
        $this->queueClient->set($queueName, $data);
    }

    public function take(string $queueName)
    {
        return $this->queueClient->get($queueName);
    }

    public function delete(string $queueName, string $taskId): void
    {
    }
}