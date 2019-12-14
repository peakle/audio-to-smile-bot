<?php

declare(strict_types=1);

namespace App\Service;

use App\Entity\User;
use Doctrine\ORM\EntityManagerInterface;
use Predis\Client as Redis;

class QueueService
{
    /**
     * @var Redis
     */
    protected $queueClient;

    /**
     * @var EntityManagerInterface
     */
    protected $manager;

    /**
     * @param EntityManagerInterface $manager
     * @param string $redisHost
     * @param string $redisPort
     */
    public function __construct(EntityManagerInterface $manager, string $redisHost, string $redisPort)
    {
        $this->queueClient = new Redis([
            'scheme' => 'tcp',
            'host' => $redisHost,
            'port' => $redisPort
        ]);

        $this->manager = $manager;
    }

    /**
     * @param string $queueName
     * @param string $data
     */
    public function put(string $queueName, string $data): void
    {
        $this->queueClient->rpush($queueName, [$data]);
    }

    /**
     * @param string $queueName
     * @return string
     */
    public function take(string $queueName): string
    {
        return $this->queueClient->lpop($queueName);
    }
}
