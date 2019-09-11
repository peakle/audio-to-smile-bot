<?php

declare(strict_types=1);

namespace App\Controller;

use App\Service\QueueService;
use Exception;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;

class DefaultController extends AbstractController
{
    private const QUEUE_NAME_SMILE = 'queue_create';

    /**
     * @var QueueService $queueService
     */
    protected $queueService;

    /**
     * @var string
     */
    protected $groupId;

    /**
     * @var string
     */
    protected $vkConfirmationToken;

    /**
     * @param QueueService $queueService
     * @param string $groupId
     * @param string $vkConfirmationToken
     */
    public function __construct(QueueService $queueService, string $groupId, string $vkConfirmationToken)
    {
        $this->queueService = $queueService;
        $this->groupId = $groupId;
        $this->vkConfirmationToken = $vkConfirmationToken;
    }

    public function index(): void
    {
        fastcgi_finish_request();
    }

    public function messageBox(Request $request): void
    {
        echo 'ok';
        fastcgi_finish_request();

        try {
            $data = json_decode($request->request, true);
            $queueBody = json_encode([
                'user_id' => $data->user_id,
                'message' => $data->body->message
            ]);

            $this->queueService->put(self::QUEUE_NAME_SMILE, $queueBody);
        } catch (Exception $exception) {
        }
    }

    public function confirmation(Request $request): ?Response
    {
        $data = json_decode($request->request, true);

        if (isset($data->type) && $data->type === 'confirmation') {
            if (isset($data->group_id) && $data->group_id === $this->groupId) {
                return new Response($this->vkConfirmationToken);
            }
        }

        fastcgi_finish_request();
    }
}