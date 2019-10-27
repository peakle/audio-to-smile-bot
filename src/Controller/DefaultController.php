<?php

declare(strict_types=1);

namespace App\Controller;

use App\Service\QueueService;
use Exception;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpClient\HttpClient;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Contracts\HttpClient\Exception\TransportExceptionInterface;

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
     *
     * @throws TransportExceptionInterface
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
     *
     * @throws TransportExceptionInterface
     */
    private function messageBox(array $data): ?Response
    {
        try {
            if (mb_strlen($data['object']['text']) === 4) {
                $message = mb_strtolower($data['object']['text']);
                if ($message === 'пиво' || $message === 'beer') {
                    $this->sendBeer($data);
                    return null;
                }
            }

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
        if ($data['group_id'] && (string)$data['group_id'] === $this->groupId) {
            return new Response($this->vkConfirmationToken);
        }

        return null;
    }

    /**
     * @param array $data
     *
     * @throws Exception
     * @throws TransportExceptionInterface
     */
    private function sendBeer(array $data): void
    {
        $userId = (string)$data['object']['from_id'];
        if (!$this->queueService->checkUserId($userId)) {
            $this->queueService->addUserId($userId);
        }
        $rand = random_int(1, 10000000);
        $res = [
            'user_id' => $userId,
            'random_id' => $rand,
            'v' => getenv('VK_API_VERSION'),
            'access_token' => getenv('VK_TOKEN'),
            'message' => 'Там в Вавилоне есть библиотека. Здесь ты найдешь ответ на свой вопрос.' .
                "\n" .
                '4,2,11,316 ' . "\n" .
                '11ts3ty8jluudzzwdlraih1azzpue8oyiynsyjuulr50ryd0qyaaqkqumfxgrpb2ino9vyxjmseytq5nwrcinqb5ob3ovyinup4xnobpoik5fjlu8jj7dka8rn1jru1d29ibey3jrhmh41ze7za99bziqys6w8m54w696js30z2u2paig92i87vzvbb8oc9nr93zcrt30mcopziutzrgf43acwg0tuy3uxym21sf61ipsgpej0pgzj8qdabj8bv615fex470zhy0cvkttn1ony95wv8h0cfv85zpngv7b4747yolxptqhvf8z4fkjwds7opjmd4lbqdz7p7hnrriolncpoi4z7md20bzy6llc9u427vml4o21javb2hls1ommz81y2tzpycxdb5i6zfjoaayd538lm7vbvtexxae7od4mkpnunbkgpaig6rfvbvizl85f558q6lcywgewgk57u07cdlwqdawzwr8a5td1ptpzvb6srr7nm8hre773m4l7zccz7h1szmibhsqce86r7bklbfsnh6ylkowvb2293j3vx22s6by31rkmbykvehqn5rjzqytncpg7fwrkhks9eniaoq9z8uvwpnnmn5ke9v66luihnvm56a97a0rqdby6qemyhhw7ezpzhda55aocojsgvc6lsdc2s9nrwnj1jght5k9clvinqlbx9aeo19d4ov1m4lplc6uph90mq64obkwijni3ni786kgso5fw101p96cupbpw1yb01dhs40y5mti4k2rvyrujcqkw0n3i5ufd2kj85o5iwlm6ltc4x3gkfl1i3va8eb2t55k91zo3hpbk92gwky93v17ftpxneb4sd8tvtzkju223tzvktkswpn3p9wjndq14dsp4hdj7xmtqybebauyehd55petmtv8d0b7tbarjxpl7mqugdupzvfp131y3odpawhlkehgszbwza41im9abgiwrcjpg7bp9wk4i4he6zhh4lut1jdyuhw5vc7y7eapgoksuxauyyv1lvq4a4k31ix5rwo46ajh5i6kimvjgdo03xca5ahrtt7h3rye5y763mqnpitvjqta00c6w84qvxb1h3di7s36rcjvik9lorvw6u23jnr2g4mnf5gosurloie3emp6mje3b7vfspfmeschyxp9qti0ns9053hyjt54lj4cyh2g2i0golzxysa0sqsh91c0lsic2hx6n72a9uvz4n11wa2xdy2kqobd2na2rmim411hfikyrji05zu9hhbworn3njwj4r1gk8l9fsp86kp5af5dckfli6rfg9tcmc9ml24n4zyo2shbq710k7hynggx9wthass8fw41rvjhr6izmkh226v2kk811j4wn3pi16ro7mcxm8txjo5y8ai0jgvwtu62lai6wfle21gfnrp4lvgl2xt4c86yztengv66st5chr4wy0go82mcyvpzo7x9c3dfllatv925pw6diu9s5g3z8wn52ykplhnfq2na11hw9yv61gu7vkorfrjjxx25lm5e8lgw19jyz3s8ag38vq3bmxil1v3sxeygfnuxy62w1f53i46q4lz6chpc408svym5024hvd763e05dd22f267frj2megqwa673wkwv0brz8qwhsegmudw7lgkhgr49ikenpewb0xj02pxgmcwpj6kowzhomjc4p2whf42do6ljffxl5hix3ngdzd8beqzoeozm3ibc33izbp6kzl6xvndahn5ti8zrhq2iww20sy5clj0inxf0rjfryw88xpxfiwpl4zasmcy3pvo0m4qlxmmjzppcqjaeh9k2pmom8l0pys8qswfjdi78n3t3tafzy1eln27d1r1vx2y4dkhv0780uqq3ip29fi2l8h75uoloofdpxp8n377tp7vmfycixznyofsrcem6sahodetcybgl0wx9yxzfzvt0vtz3h393uzuh1lvke6ntmzukdxqvp4qb5iva44h3udtcpc32sx21n3b3kp8fg6eufx3qu5huhtjk0vs60cxqfaxxu9opizve9pi80r9y2qfhp3glpgeusk611xoh778vhs71pek0xddpx6b5uupisg9iigbc1ucvfltacd9imal0z7jc1ioivzhlhdi4199uzg2kyg93wq4fv4tsennkheocjcq7xh2jmngsn3qz1ym6hqzlsztoc76gdgodxp9v36jcc4bm66ngxeywsgkoi1iytug3mecy5u6syygo4qgx23x6r3gnyqe3rpc4hgy2708ia4u9ku5nlj568w6dwo69mrs1f1cw1wy986wx5l5p76zsmw0rduks698tr5cmp1orkl5gev8765iu88np3z463j3hy9q2u4hx542ipmgssaaqjpmbh8w428xu44hbs0x06ea9yi5ibdls6rtpvwx67v6qj45lu5rnlspb9y3w3xoi2x35p5s1jlf67ktcnmlhv4n5l5gmgekkx3tw2lcu1n7gmhongd59wtc2k7ot4bsqppbopa68eje9captqttzjxylxto8mkohpjkfvip05mye598xjpso12wh5tj1cvxmjr1r6yk6y8tedzorek5cv94gupa6g5qw6r7wo0v309f284vkj5zjbudjafvzmw0tjbcpr5samsc4gjk0vcs9hdy9gk4gw9j5ojkle0407wofts58byt31cfihript1tamdq6yo9sqbi1h0ldor10h5zwbu7tg8tx2henc1csy1c68b2zwa65rxgberaerpm1y3o1t349lf112jhzkmg7owl6ej975fyj5dis6vm5pwyuaewjkgyu43mtw97g2jqrvvtscbbilkmlt6pmljp4jvtufjoon0qbe56v01q6kauyhsivfx9ha5ekne5pjurho25g0arawrx90dwjher4aqw2qfaffbe220e8g0eju0elrnn4hvzhcsn95ijznvsym386cavqvcknolss9opfibcdni8zqnsc35sdiu2erbmgnx3i0dhnyqbpr364twddg9pqx9ho492w6rf7mb9e199w067wp10beulb3tv3n9n94hr6bz83c3g8t9a47smar31tdqgxrla1f2qhfjtjpak6dkan3yaqgg5d4fauqcwituyuhz8dcklqiaz0cfzu4eqkclbjsdr6l95nwhlabczx74ifghainwl9grl46vbwfnaqljej1qsngj9wtao900ak9d7robtekftw8gwywjxk311w2w83q'
        ];

        echo 'ok';

        $httpClient = HttpClient::create();
        $httpClient->request('POST',
            'https://api.vk.com/method/messages.send',
            [
                'body' => $res
            ]
        );

        if (!$this->queueService->checkUserId($userId)) {
            $this->queueService->addUserId($userId);
        }
    }
}
