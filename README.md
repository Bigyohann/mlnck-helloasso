# 🔗 HelloAsso Events API (Go) — MLNCK

[![codecov](https://codecov.io/gh/Bigyohann/mlnck-helloasso/graph/badge.svg?token=JRd1UWxiGr)](https://codecov.io/gh/Bigyohann/mlnck-helloasso)
[![Tests](https://github.com/Bigyohann/mlnck-helloasso/actions/workflows/test.yml/badge.svg)](https://github.com/Bigyohann/mlnck-helloasso/actions/workflows/test.yml)

Une petite API en **Go** pour connecter le site [mlnck.fr](https://mlnck.fr) à [HelloAsso](https://www.helloasso.com/), et exposer les **prochains événements publics**.

---

## ✨ Objectif

Permettre au site MLNCK d'afficher dynamiquement les événements à venir créés sur HelloAsso.

---

## ⚙️ Fonctionnalités

- Authentification OAuth2 avec HelloAsso
- Récupération des événements à venir
- Mise en cache simple pour éviter de surcharger l'API HelloAsso

---

## 🚀 Lancer l'API

### 1. Cloner le projet

```bash
git clone https://github.com/<utilisateur>/mlnck-helloasso-api-go.git
cd mlnck-helloasso-api-go
```

### 2. Configurer l'application
Créer un fichier .env (ou exporter les variables d’environnement) :

```dotenv
CLIENT_ID=your_client_id
CLIENT_SECRET=your_client_secret
ORGANIZATION=mlnck
PORT=8080
```

### 3. Lancer le projet
```bash
go build .
./mlnck
```

### 4. Tester l'API
L'API sera disponible sur : http://localhost:8080/events

## 🧪 Exemple de réponse
```json
[
	{
		"banner": {
			"fileName": "croppedimage-a74239b9dd4b4aba887e4779d474ee10.png",
			"publicUrl": "https://cdn.helloasso.com/img/photos/evenements/croppedimage-a74239b9dd4b4aba887e4779d474ee10.png"
		},
		"currency": "EUR",
		"description": "Le club Kayak vous invite à vivre une journée d’évasion en kayak autour des magnifiques Îles de Lérins. Le départ s’effectuera en groupe depuis le parking du Palm Beach à Cannes, avec un moniteur. Vous pagaierez en direction des îles pour une exploration conviviale entre mer et nature.\nUne pause pique-nique est prévue sur l’une des plages de l’archipel, suivie d’un moment de détente. Le retour se fera ensuite tranquillement vers le point de départ.\n",
		"startDate": "2025-07-05T10:30:00+02:00",
		"endDate": "2025-07-05T16:30:59+02:00",
		"meta": {
			"createdAt": "2025-06-14T17:13:41.133+02:00",
			"updatedAt": "2025-06-24T17:39:07.49+02:00"
		},
		"title": "Îles de Lérins - Journée",
		"formSlug": "iles-de-lerins-journee",
		"url": "https://www.helloasso.com/associations/mandelieu-la-napoule-canoe-kayak/evenements/iles-de-lerins-journee"
	}
]
```
