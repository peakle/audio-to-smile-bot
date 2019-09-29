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
                'user_id' => $data['object']['from_id'],
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
        $rand = random_int(1, 10000000);
        $res = [
            'user_id' => $data['object']['from_id'],
            'random_id' => $rand,
            'v' => getenv('VK_API_VERSION'),
            'access_token' => getenv('VK_TOKEN'),
            'message' => 'Там в Вавилоне есть библиотека. Здесь ты найдешь ответ на свой вопрос.' .
                "\n" .
                'Стена 1 стеллаж 5 Том 14 страница 180 ' . "\n" .
                '0u9vnzkp7ylbzmxqbp2yxoe6axq0v9qfpbcs9j5rxwtgvf1qu6vh63c0fp4xhgu0snulq3qiccbkuaezz6wnb5pprw4410yupyle89t9n908ro9apj70ygg3vaycz97bky0f7sq0cyybde0ncbci9pskykgttg6j2lu7ivwl6iqxwczu7nyfsncwnhwc3aeld6vhnzq938f7c1xhna5hf11e81b6u7cmpcpwuwp8yjmgclj4m5t5cr4m1cc7zjj6w6oygp562mtery1jkqbpct6h8f57rww8igh872yrjx1yljmk8jupuuha76w4ndcfid7j9leo13r0djx4o14twe8zjf6jvbk65wl0zwszkpuk0913tzal0uktngfpk45iwlna6g2ccai0rp6ntrfk3nc3g3u90ad8omeoohkwp4c50iiub7tear9g1chhk25hk7hl1tqsabc2m1me7lkqzhlwffknp78ofzupu156rbyutjj02nb9vfltn9mzjj88ckzwro626jht5lj1geivfmg5uhvurji3zpydby9q57cpcvceir8v9ccnou5411rdcpdt0d23eg6n7pwlfp6whzddthbj9jpi8yf0ltc66yf7kvfib6fl1vv2ttsszjo1f4cjt8w0llle9ztfdzzxjfppnavm1jkme3qfb9n5wdeshjtc5lw6zazccg5vckxz4i1mn3k99k9ui7dpo0r0qp97fodot2leiooh8k8z7mxqwwnnscmjwsa4uz8uokkzzrcosymi1t9k515przi74gxnxa7kavbkizbvrcx3i8w2mce1qyfdic6fx3inw3cx48ezu1xwopttj77saqs72z342nyxcjmql42b6lnwewg67qhf6qcc4n5kkhm5jb23kaywii6w2h1c60lo7xu4ksmjecqxurwl84snim02y03i7z9okg7yto6ns8hdmbh2q7931sqwdldjdsrm5lqxwh4c767n2c7jhn7g332ehyzfa6l7j139z2f8zrv91xb88z3tp4hi4xpxw0ofoeynaspyvqehwa3qkfp5mg5i4nr9xa3opcndgig4r0lesuylyxk82r0tgwrwlygz71kxzwcwowg13iiqwjljgm7ylsae6g6umrzwmvl3fv7oaag8nvzjuzfnu85d0q3tvpk9t7i8bwroq0hxluey8wb9klg44hx71uejkwnuqkadsbwguahvphqkawjfrj3kk7yadebbeo3v4u6wlfpgorsachig5vqm0n56lo1ossfhhdzv2qa3cn1aawijce111hnkeytlm1k3ppwbjh333qma31n4hajdksy2kt8rwlz7q7gcalo3k8dk7fhw9xteex45cg7oxubns51jogg82nsjvfyds2lctxvt05h1wwi93860cz6b1y102f2qizqz9q6zuaf9ceqsr92hj9wht9m2nvsrndjrlyfz9nqfh1hnat2eqzo7jlfkr1tzyk6x05h7cgpumkkca0fmllc2uz9gbybcy8n1z20e4au795ky777zmv11ee5k08fr9ijjjoxvgt0ioub8o15wr1dsl3fcombkxxna9upcrde33xt4jroulyp4k49774wzlnax22hdjk39ie8m3v5pmdvm8am89l9lld1km0uy5g7p7ma47a3mab7ar3nazn74kzrnxextgm675y2xcygmithwgh1nerm00mwb1ul1rtgdiiwwrejd4hkcfltyabdggloqiuf82qnovsqkzv9ykb17762yv7dui4z94bjubzpgdf5o8qao7cueihuocf3qv36ox44al49g8zh9ce4stsf7k964cdge4o19sdnoxdn6is4wrifgvcucam1f0lo0iozx6cd690ne7bc8keh0g3w0z9dnnpj51o1r2lhapqmqrbp3jb7abce13fcn42qzl4ibyku4z6c5zedydpsi4ftm3lcoqqzp050zu2qvvwv285nzzv575n1snahf76t52oujpuq32lbhj0g8rywk1w9tqrekl9f5qfy2dpu6545opcs9bnvwwz5pqj1nmumru8kj6gpexqhhezkisj9h25oevoc1hncckpl0f3xbms5hpln37hf7zf3ro5na8upgyr54x6k28tz02i6aw0884j3lbc9gmnkbc1waanycfgz6c1uijqisdwaik1asse118v3nrbdjit2gzmcl6kxblzfcoibkqtt1zb3kpr4sxjfmoomkgj12sjdp3mfsg7jea85gotvnzp5qvqtlq9oib675mop290g7t7d30jfcz68uamuzcfxc0qliewmrrp3sr3dvgdxn9xocf0gx7fs76b3yfvy4pcqlgz6pn6079j6wcxp7ciow3ujtej7zwxi0cci1txxdl1laq2s2ag0j6byl1ccfqw4p5z42wj1bsqphpy6nk25vcy0mskey4lwuc8nicbw9wx0op3ejpb52c3p4nuxc023fvp8jq945x6v1733apm9ykioaqb07hvfx2shgdrpwiw6k06sjad7fn0dyl9xg69eycq3eunbk79bil2oung4yiu6g0ra5157vm7x4bgmpoh82fdgwrw357s65v0buwu5eni5bwe20mot6yecodxuwcwz87fytszdd81i8016r56xl6b7mnrbz8gevf2y3klgneoihtcwazdac2zanj8yw5u13e0oeka4k2kgczqrthsd6ctx61cnoojkotiyypteoqxn2r2qi6go5l72khhjpif0wln3sku8vx9bxwvxra6xjxm196srwku06ry25oaoorftizas9hfkygp8b5rm1ynwv6f6thgqg67f3ndacmk78hq9ouiqkpakhpoukmbxgk3uqmsc89ivgu92xw06iq44p1cpkmof2g0dbea7v208ja2u7vb5407t1a9w9srjcwyfsdrm7rxjbgex7fx29xlp2brtidrtewkycfazf8gz7jllcx7j3du2zav9vrhh4avyotlo2dmmn7dighvbs2rorafnqkrzykmfrf20dsld9td83ntnioeomy50wl4hcugzz98oplkqt6go238waiw4j6wjszngk0pisjpwizsyoia7mhkf9xepxqm131yke6zaepegxa7curev6f80hligej6efanmdp9sjdhffztrc4iofsbbde52y8nwzbgd3r6st0usqnfps0r5qof8oql3atqxa349lybw3o5xeiogle3faeas71878wi37x2z7cyeeqo6eqd'
        ];

        echo 'ok';

        $httpClient = HttpClient::create();
        $httpClient->request('POST',
            'https://api.vk.com/method/messages.send',
            [
                'body' => $res
            ]
        );
    }
}
