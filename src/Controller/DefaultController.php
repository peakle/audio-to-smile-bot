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
     * @var string
     */
    private $secret;

    /**
     * @param QueueService $queueService
     * @param string $groupId
     * @param string $vkConfirmationToken
     * @param string $secret
     */
    public function __construct(
        QueueService $queueService,
        string $groupId,
        string $vkConfirmationToken,
        string $secret
    ) {
        $this->queueService = $queueService;
        $this->groupId = $groupId;
        $this->vkConfirmationToken = $vkConfirmationToken;
        $this->secret = $secret;
    }

    /**
     * @param Request $request
     *
     * @return Response|null
     */
    public function index(Request $request): ?Response
    {
        $content = $request->getContent();
        $data = json_decode($content, true);

        if (!isset($data['secret']) || $data['secret'] !== $this->secret) {
            return $this->render('base.html.twig');
        }

        if (isset($data['type'])) {
            switch ($data['type']) {
                case 'confirmation':
                    return $this->confirmation($data);
                case 'message_new':
                    return $this->messageBox($data);
                default:
                    return null;
            }
        }

        return null;
    }

    /**
     * @param array $data
     *
     * @return Response|null
     */
    private function messageBox(array $data): ?Response
    {
        try {
            $queueBody = json_encode([
                'user_id' => (string)$data['object']['from_id'],
                'message' => $data['object']['text']
            ]);

            $this->queueService->put(self::QUEUE_NAME_SMILE, $queueBody);
            return new Response('ok');
        } catch (Exception $exception) {
            fastcgi_finish_request();
        }

        return null;
    }

    /**
     * @param array $data
     *
     * @return Response|null
     */
    private function confirmation(array $data): ?Response
    {
        if (isset($data['group_id']) && (string)$data['group_id'] === $this->groupId) {
            return new Response($this->vkConfirmationToken);
        }

        return null;
    }
}
