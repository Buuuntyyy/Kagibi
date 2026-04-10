<!-- Copyright (C) 2025-2026  Buuuntyyy -->
<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

<template>
  <div class="legal-container">
    <div class="legal-content">
      <h1>Politique de Confidentialité</h1>
      <p class="last-updated">Dernière mise à jour : 8 Janvier 2026</p>

      <section>
        <h2>1. Introduction</h2>
        <p>Dans le cadre de son activité, Kagibi est amené à traiter des informations vous concernant. La présente Politique de Confidentialité vous informe de la manière dont nous recueillons et traitons vos données personnelles. Nous vous invitons à la lire attentivement.</p>
        <p>Nous attachons la plus grande importance au respect de la vie privée et des données individuelles de nos utilisateurs. Nous nous engageons à respecter les dispositions de la loi n°78-17 du 6 janvier 1978 relative à l'informatique, aux fichiers et aux libertés modifiée et au Règlement (UE) 2016/679 du Parlement européen et du Conseil du 27 avril 2016 dit « <em>RGPD</em> ».</p>
      </section>

      <section>
        <h2>2. Responsable de traitement</h2>
        <p>Le responsable du traitement des données est [NOM DU RESPONSABLE OU DE L'ENTREPRISE], joignable à l'adresse suivante : [ADRESSE] ou par email à [EMAIL DE CONTACT].</p>
      </section>

      <section>
        <h2>3. Données collectées</h2>
        <p>Nous collectons et traitons les données suivantes :</p>
        <ul>
          <li><strong>Données d'identification :</strong> Nom d'utilisateur, adresse email.</li>
          <li><strong>Données techniques :</strong> Logs de connexion, adresse IP, type de navigateur (à des fins de sécurité).</li>
          <li><strong>Contenu utilisateur :</strong> Fichiers et dossiers que vous stockez sur le service. Ces fichiers sont chiffrés.</li>
        </ul>
      </section>

      <section>
        <h2>4. Finalités du traitement</h2>
        <p>Vos données sont collectées pour les finalités suivantes :</p>
        <ul>
          <li>Fourniture et gestion du service Kagibi (création de compte, stockage, partage).</li>
          <li>Sécurité du service et prévention de la fraude.</li>
          <li>Communication avec les utilisateurs (support technique, notifications importantes).</li>
          <li>Respect des obligations légales et réglementaires.</li>
        </ul>
      </section>

      <section>
        <h2>5. Destinataires des données</h2>
        <p>Les données collectées sont destinées aux équipes techniques de Kagibi. Elles peuvent être transmises aux sous-traitants auxquels Kagibi fait appel dans le cadre de l'exécution de ses services (ex: hébergeur).</p>
        <p>Kagibi ne vend ni ne loue vos données personnelles à des tiers à des fins de marketing.</p>
      </section>

      <section>
        <h2>6. Durée de conservation</h2>
        <p>Vos données sont conservées pendant toute la durée de l'utilisation du service (tant que votre compte est actif).</p>
        <p>En cas de suppression de compte, vos données personnelles et fichiers seront supprimés de nos serveurs actifs sous un délai de 30 jours. Certaines données (logs) peuvent être conservées plus longtemps pour répondre à des obligations légales, pour une durée n'excédant pas celle prévue par la loi.</p>
      </section>

      <section>
        <h2>7. Vos droits</h2>
        <p>Conformément à la réglementation applicable, vous disposez des droits suivants :</p>
        <ul>
          <li><strong>Droit d'accès :</strong> Vous pouvez demander l'accès à vos données personnelles.</li>
          <li><strong>Droit de rectification :</strong> Vous pouvez demander la correction de données inexactes ou incomplètes.</li>
          <li><strong>Droit à l'effacement :</strong> Vous pouvez demander la suppression de votre compte et de vos données ("droit à l'oubli").</li>
          <li><strong>Droit à la limitation du traitement :</strong> Vous pouvez demander de geler l'utilisation de vos données.</li>
          <li><strong>Droit à la portabilité :</strong> Vous pouvez récupérer vos données dans un format structuré.</li>
          <li><strong>Droit d'opposition :</strong> Vous pouvez vous opposer au traitement de vos données pour des motifs légitimes.</li>
          <li><strong>Sort des données après la mort :</strong> Conformément à la loi "Informatique et Libertés", vous pouvez définir des directives relatives à la conservation, à l'effacement et à la communication de vos données après votre décès.</li>
        </ul>
        <p>Pour exercer ces droits, vous pouvez nous contacter à l'adresse : [EMAIL DE CONTACT].</p>
        <p>Si vous estimez, après nous avoir contactés, que vos droits "Informatique et Libertés" ne sont pas respectés, vous pouvez adresser une réclamation à la CNIL.</p>
      </section>

      <section>
        <h2>8. Sécurité des données</h2>
        <p>Kagibi met en œuvre des mesures techniques et organisationnelles appropriées pour protéger vos données contre l'accès non autorisé, la modification, la divulgation ou la destruction.</p>
      </section>

      <section>
        <h2>9. Cookies</h2>
        <p>Le site utilise des cookies techniques strictement nécessaires au fonctionnement du service (maintien de la session connecté). Ces cookies ne nécessitent pas de consentement préalable.</p>
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
