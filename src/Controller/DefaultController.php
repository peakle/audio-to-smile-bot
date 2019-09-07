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
     * @param QueueService $queueService
     */
    public function __construct(QueueService $queueService)
    {
        $this->queueService = $queueService;
    }

    public function index(): void
    {
        fastcgi_finish_request();
    }

    public function messageBox(Request $request): Response
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
            $groupId = $this->getParameter('group_id');

            if (isset($data->group_id) && $data->group_id === $groupId) {
                $token = $this->getParameter('vk_confirmation_token');
                return new Response($token);
            }
        }

        fastcgi_finish_request();
    }
}