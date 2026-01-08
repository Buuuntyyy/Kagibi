<template>
  <div class="legal-container">
    <div class="legal-content">
      <h1>Conditions Générales d'Utilisation (CGU)</h1>
      <p class="last-updated">Dernière mise à jour : 8 Janvier 2026</p>

      <section>
        <h2>1. Objet</h2>
        <p>Les présentes Conditions Générales d'Utilisation (ci-après "CGU") ont pour objet de définir les modalités de mise à disposition des services du site et de l'application <strong>SaferCloud</strong>, ci-après nommé "le Service", et les conditions d'utilisation du Service par l'Utilisateur.</p>
      </section>

      <section>
        <h2>2. Mentions Légales</h2>
        <p>L'édition du site SaferCloud est assurée par [NOM DE L'ENTREPRISE OU DE L'INDIVIDU], dont le siège social est situé à [ADRESSE COMPLETE].</p>
        <p>Adresse e-mail de contact : [EMAIL DE CONTACT]</p>
        <p>L'hébergeur du Service est : [NOM DE L'HEBERGEUR, ex: AWS, OVH, Google Cloud], situé à [ADRESSE DE L'HEBERGEUR].</p>
      </section>

      <section>
        <h2>3. Accès au Service</h2>
        <p>Le Service permet à l'Utilisateur un accès gratuit ou payant aux services suivants :</p>
        <ul>
          <li>Stockage de fichiers en ligne (Cloud) ;</li>
          <li>Partage de fichiers et de dossiers via des liens sécurisés ;</li>
          <li>Chiffrement des données.</li>
        </ul>
        <p>Le Service est accessible gratuitement en tout lieu à tout Utilisateur ayant un accès à Internet. Tous les frais supportés pour l'accès au Service (matériel informatique, logiciels, connexion Internet, etc.) sont à la charge de l'Utilisateur.</p>
      </section>

      <section>
        <h2>4. Responsabilité de l'Utilisateur</h2>
        <p>L'Utilisateur est responsable des risques liés à l'utilisation de son identifiant de connexion et de son mot de passe. Le mot de passe de l'Utilisateur doit rester secret. En cas de divulgation de mot de passe, l'Éditeur décline toute responsabilité.</p>
        <p>L'Utilisateur assume l'entière responsabilité de l'utilisation qu'il fait des informations et contenus présents sur le site SaferCloud.</p>
        <p>Tout usage du service par l'Utilisateur ayant directement ou indirectement pour conséquence des dommages doit faire l'objet d'une indemnisation au profit du site.</p>
        <p>Le membre s'engage à ne pas mettre en ligne de contenus pouvant porter atteinte aux intérêts de tierces personnes. L'Utilisateur s'interdit de stocker ou partager des fichiers illégaux, contrefaits, ou contraires aux bonnes mœurs.</p>
      </section>

      <section>
        <h2>5. Responsabilité de l'Éditeur</h2>
        <p>Tout dysfonctionnement du serveur ou du réseau ne peut engager la responsabilité de l'Éditeur.</p>
        <p>De même, la responsabilité du site ne peut être engagée en cas de force majeure ou du fait imprévisible et insurmontable d'un tiers.</p>
        <p>Le site SaferCloud s'engage à mettre en œuvre tous les moyens nécessaires pour garantir la sécurité et la confidentialité des données. Toutefois, il ne peut garantir une sécurité absolue.</p>
      </section>

      <section>
        <h2>6. Propriété Intellectuelle</h2>
        <p>Les marques, logos, signes ainsi que tous les contenus du site (textes, images, son...) font l'objet d'une protection par le Code de la propriété intellectuelle et plus particulièrement par le droit d'auteur.</p>
        <p>L'Utilisateur doit solliciter l'autorisation préalable du site pour toute reproduction, publication, copie des différents contenus. Il s'engage à une utilisation des contenus du site dans un cadre strictement privé, toute utilisation à des fins commerciales et publicitaires est strictement interdite.</p>
      </section>

       <section>
        <h2>7. Données Personnelles</h2>
        <p>L'Utilisateur dispose de droits d'opposition, d'accès, de rectification et de suppression portant sur les données personnelles le concernant, conformément à la Politique de Confidentialité du site.</p>
      </section>

      <section>
        <h2>8. Évolution des CGU</h2>
        <p>Le site SaferCloud se réserve le droit de modifier les clauses de ces conditions générales d'utilisation à tout moment et sans justification.</p>
      </section>

      <section>
        <h2>9. Droit applicable, médiation et juridiction compétente</h2>
        <p>La législation française s'applique au présent contrat. En cas d'absence de résolution amiable d'un litige né entre les parties, les tribunaux français seront seuls compétents pour en connaître.</p>
        <p>Pour tout litige de consommation qui n'aurait pu être résolu directement avec l'Éditeur, le consommateur peut recourir gratuitement à un médiateur de la consommation en vue de la résolution amiable du litige (Code de la consommation art. L.612-1). Les coordonnées du médiateur de référence sont : [NOM ET COORDONNÉES DU MÉDIATEUR].</p>
        <p>La Commission Européenne met à disposition une plateforme de règlement en ligne des litiges (RLL) : <a href="https://ec.europa.eu/consumers/odr" target="_blank">https://ec.europa.eu/consumers/odr</a>.</p>
      </section>
      
      <button @click="$router.go(-1)" class="btn-back">Retour</button>
    </div>
  </div>
</template>

<style scoped>
.legal-container {
  display: flex;
  justify-content: center;
  padding: 40px 20px;
  background-color: var(--background-color, #f5f7fa);
  min-height: 100vh;
  font-family: 'Segoe UI', sans-serif;
  color: #333;
}

.legal-content {
  background: white;
  padding: 40px;
  border-radius: 8px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.05);
  max-width: 800px;
  width: 100%;
}

h1 {
  color: var(--primary-color, #42b983);
  margin-bottom: 10px;
}

h2 {
  font-size: 1.4rem;
  margin-top: 30px;
  margin-bottom: 15px;
  color: #2c3e50;
  border-bottom: 2px solid #f0f0f0;
  padding-bottom: 10px;
}

p, li {
  line-height: 1.6;
  color: #555;
  margin-bottom: 15px;
}

ul {
  padding-left: 20px;
  margin-bottom: 15px;
}

.last-updated {
  font-style: italic;
  color: #888;
  margin-bottom: 30px;
  border-bottom: 1px solid #eee;
  padding-bottom: 20px;
}

.btn-back {
  margin-top: 30px;
  padding: 10px 20px;
  background-color: var(--primary-color, #42b983);
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
}

.btn-back:hover {
  opacity: 0.9;
}
</style>
