<?php

namespace App\DataFixtures;

use App\Entity\Smile;
use Doctrine\Bundle\FixturesBundle\Fixture;
use Doctrine\Common\Persistence\ObjectManager;

class SmileFixtures extends Fixture
{
    public function load(ObjectManager $manager)
    {
        $smiles = [
            128512 => [
                'sample' => 'happy.ogg',
                'comment' => '😀',
            ],

            128514 => [
                'sample' => 'laugh.ogg',
                'comment' => '😂'
            ],

            128529 => [
                'sample' => 'sss.ogg',
                'comment' => '😑'
            ],
        ];

        foreach ($smiles as $code => $smile) {
            $newSmile = new Smile();
            $newSmile->setCode($code);
            $newSmile->setSample($smile['sample']);

            $manager->persist($newSmile);
        }
        $manager->flush();
    }
}
