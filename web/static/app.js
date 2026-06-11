'use strict';

/* ─── Internationalisation ───────────────────────────────────────────────── */
const TRANSLATIONS = {
  fr: {
    // Clock
    days:   ['Dimanche','Lundi','Mardi','Mercredi','Jeudi','Vendredi','Samedi'],
    months: ['janvier','février','mars','avril','mai','juin','juillet','août','septembre','octobre','novembre','décembre'],

    // Login page
    'login.email':              'Email',
    'login.password':           'Mot de passe',
    'login.totp':               'Code TOTP',
    'login.submit':             'Connexion',
    'login.forgot':             'Mot de passe oublié ?',
    'login.register':           'Pas encore de compte ?',
    'login.sso':                'Se connecter avec',
    'login.or':                 'ou',
    'forgot.back':              'Retour',
    'forgot.hint':              'Entrez votre email, vous recevrez un lien de réinitialisation.',
    'forgot.submit':            'Envoyer le lien',
    'forgot.sent':              "Si l'adresse existe, un email a été envoyé.",
    'register.username':        'Nom d\'utilisateur',
    'register.email':           'Email',
    'register.password':        'Mot de passe',
    'register.password.hint':   'min. 12 caractères',
    'register.hint':            'Saisissez vos informations — vous recevrez un lien de configuration par email.',
    'register.submit':          'Demander un compte',
    'register.sent':            'Lien envoyé ! Vérifiez votre boîte mail.',
    'register.back':            'Retour',
    'setup.title':              'Configuration du compte',
    'setup.totp.hint':          'Scannez avec votre application d\'authentification',
    'setup.next':               'Continuer',
    'setup.scan':               'Valider et continuer',
    'setup.finish':             'Créer le compte',
    'setup.invalid':            'Ce lien est invalide ou a expiré.',
    'setup.password.hint':      'Choisissez un mot de passe fort (min. 12 caractères)',
    'setup.password.confirm':   'Confirmer le mot de passe',
    'setup.password.mismatch':  'Les mots de passe ne correspondent pas.',
    'setup.open.app':           'Ouvrir dans l\'appli d\'authentification',
    'btn.copy':                 'Copier',
    'account.delete':           'Supprimer le compte',
    'account.delete.desc':      'Suppression définitive — toutes vos données seront effacées (RGPD).',
    'account.delete.confirm':   'Confirmez votre mot de passe pour supprimer votre compte.',
    'account.delete.btn':       'Supprimer définitivement',
    'admin.delete.warning':     'Suppression définitive — toutes les données de cet utilisateur seront effacées (RGPD). Action irréversible.',

    // Hub tiles
    'hub.bookmarks':         'Marque-pages',
    'hub.bookmarks.sub':     'Parcourir & rechercher',
    'hub.account':           'Compte',
    'hub.account.sub':       'Profil, sécurité',
    'hub.theme':             'Thème',
    'hub.theme.sub':         'Fonds d\'écran & effets',
    'hub.admin':             'Administration',
    'hub.admin.sub':         'Utilisateurs, réglages',
    'hub.logout':            'Déconnexion',
    'hub.logout.sub':        'Fermer la session',
    'hub.setup':             '⚙ Configuration requise — SMTP non configuré. Cliquez pour configurer.',

    // Panel titles & tabs
    'panel.bookmarks':       'Marque-pages',
    'panel.account':         'Mon compte',
    'panel.theme':           'Thème',
    'panel.admin':           'Administration',
    'tab.profile':           'Profil',
    'tab.security':          'Sécurité',
    'tab.engine':            'Recherche',
    'tab.sessions':          'Sessions',
    'tab.effects':           'Effets',
    'tab.wallpapers':        'Fonds d\'écran',
    'tab.bookmarklet':       'Bookmarklet',
    'tab.stats':             'Statistiques',
    'tab.users':             'Utilisateurs',
    'tab.invitations':       'Invitations',
    'tab.settings':          'Réglages',
    'tab.audit':             'Journal',

    // Bookmarks panel
    'bm.add':                '+ Ajouter',
    'bm.import':             'Importer',
    'bm.export':             'Exporter',
    'bm.search.placeholder': 'Filtrer les marque-pages…',
    'bm.all.folders':        'Tous les dossiers',
    'bm.modal.add':          'Ajouter un marque-page',
    'bm.modal.edit':         'Modifier le marque-page',
    'bm.modal.title':        'Titre',
    'bm.modal.title.ph':     'Titre du marque-page',
    'bm.modal.folder':       'Dossier',
    'bm.modal.folder.ph':    'Dev/Go  (laisser vide pour aucun)',
    'bm.modal.tags':         'Tags',
    'bm.modal.tags.ph':      'homelab, selfhosted  (séparés par virgule)',
    'bm.empty':              'Aucun marque-page. Utilisez + Ajouter ou Importer.',
    'bm.no.folder':          'Sans dossier',
    'bm.url.required':       'URL et titre requis.',
    'bm.confirm.delete':     'Supprimer ce marque-page ?',
    'bm.prev':               '← Précédent',
    'bm.next':               'Suivant →',
    'bm.import.ok':          'Import terminé',
    'bm.import.added':       'ajoutés',
    'bm.import.skipped':     'ignorés',

    // Profile tab
    'profile.stats':         'Statistiques',
    'profile.title':         'Profil',
    'profile.username':      'Nom d\'utilisateur',
    'profile.email':         'Email',
    'profile.save':          'Enregistrer',
    'profile.locale':        'Langue de l\'interface',
    'profile.locale.fr':     'Français',
    'profile.locale.en':     'English',

    // Security tab
    'security.password':          'Changer le mot de passe',
    'security.current':           'Mot de passe actuel',
    'security.new':               'Nouveau mot de passe',
    'security.change':            'Modifier',
    'security.changed':           'Mot de passe modifié',
    'security.totp':              'Authentification à deux facteurs',
    'security.totp.note':         'Le TOTP ajoute un code à 6 chiffres lors de la connexion (Google Authenticator, etc.)',
    'security.totp.configure':    'Configurer / Réinitialiser le TOTP',
    'security.totp.enable':       'Activer la 2FA',
    'security.totp.disable':      'Désactiver le TOTP',
    'security.totp.code':         'Code de vérification',
    'security.totp.confirm':      'Confirmer',
    'security.totp.confirm.prompt': 'Confirmez avec votre mot de passe actuel :',

    // Sessions tab
    'sessions.title':        'Sessions actives',
    'sessions.revoke':       'Révoquer',
    'sessions.revoke_all':   'Révoquer toutes les sessions',
    'sessions.current':      '(session actuelle)',
    'sessions.bookmarklet':  'Bookmarklet',
    'sessions.none':         'Aucune session.',
    'sessions.revoke.all.confirm': 'Révoquer toutes les sessions (vous serez déconnecté) ?',

    // Section titles (JS-built panels)
    'section.engine':        'Moteur de recherche',
    'section.wallpapers':    'Fonds d\'écran',
    'section.bookmarklet':   'Bookmarklet',
    'section.effects':       'Effets visuels',
    'section.registration':  'Inscription ouverte',
    'admin.reg.enable':      'Désactivée — activer',
    'admin.reg.disable':     'Activée — désactiver',
    'admin.reg.desc':        'Autoriser l\'inscription publique sans invitation',

    // Admin
    'admin.stat.users':      'Utilisateurs',
    'admin.stat.active':     'Actifs',
    'admin.stat.bookmarks':  'Marque-pages',
    'admin.stat.wallpapers': 'Fonds d\'écran',
    'admin.stat.db':         'Base de données',
    'admin.user.create':     'Créer un utilisateur',
    'admin.user.create.btn': 'Créer',
    'admin.user.temp.password': 'Mot de passe temporaire',
    'admin.user.suspend':    'Suspendre',
    'admin.user.activate':   'Activer',
    'admin.user.delete':     'Supprimer',
    'admin.user.you':        '(vous)',
    'admin.user.stats':      '… marque-pages · … fonds · … sessions',
    'admin.user.delete.confirm': 'Supprimer l\'utilisateur {u} ? Action irréversible.',
    'admin.inv.none':        'Aucune invitation.',
    'admin.inv.used':        'utilisée',
    'admin.inv.expired':     'expirée',
    'admin.inv.pending':     'en attente',
    'admin.inv.resend':      'Renvoyer',
    'admin.inv.revoke':      'Révoquer',
    'admin.inv.expires':     'expire',
    'admin.inv.invite':      'Inviter',
    'admin.menu.label':      'Menu',
    'admin.menu.desc':       'Le bang qui ouvre le menu plein écran. Tape-le dans la barre de recherche.',
    'admin.menu.locked':     'Configuré via compose (CAIRN_MENU_BANG) — non modifiable ici.',
    'admin.saved':           'Enregistré.',
    'admin.section.system':  'Système',
    'admin.section.infra':   'Infrastructure (compose · lecture seule)',
    'admin.section.smtp':    'Email (SMTP)',
    'admin.section.sso':     'SSO (OpenID Connect)',
    'admin.sso.desc':        'Connecte un fournisseur OIDC (Authentik, Keycloak…). Le bouton SSO apparaîtra sur la page de connexion.',
    'admin.upload.limit':      'Limite upload',
    'admin.upload.limit.save': 'Enregistrer la limite',
    'admin.upload.limit.reset': 'Remettre à la limite globale',
    'admin.upload.global':     'limite globale',
    'admin.quota':             'Quota stockage',
    'admin.quota.save':        'Enregistrer le quota',
    'admin.quota.reset':       'Remettre au quota global',
    'admin.limit.file':        'Fichier (Mo)',
    'admin.limit.storage':     'Stockage (Mo)',
    'admin.pending.title':     'Demandes d\'inscription en attente',
    'admin.pending.none':      'Aucune demande en attente.',
    'admin.pending.revoke':    'Révoquer',
    'admin.pending.completed': 'complétée',
    'admin.sys.addr':        'Adresse d\'écoute',
    'admin.sys.env':         'Environnement',
    'admin.sys.base_url':    'URL publique',
    'admin.sys.db_path':     'Base de données',
    'admin.sys.media_path':  'Répertoire médias',
    'admin.sys.max_upload':  'Taille max upload',
    'admin.sys.storage_quota': 'Quota stockage défaut',
    'admin.sys.trusted_proxy': 'Proxy de confiance',
    'admin.sys.session_secret': 'Secret de session',
    'admin.sys.version':     'Version',
    'admin.sys.set':         'défini',
    'admin.sys.not_set':     '⚠ non défini',
    'admin.sys.yes':         'oui',
    'admin.sys.no':          'non',

    // Engine
    'engine.custom':         'Personnalisé',
    'engine.custom.prompt':  'URL du moteur (doit finir par =) :',

    // Wallpapers
    'wp.upload.label':       '+ Cliquez ou glissez pour uploader un fond d\'écran',
    'wp.none':               'Aucun fond d\'écran.',
    'wp.pin':                'Épingler',
    'wp.unpin':              'Désépingler',
    'wp.delete.confirm':     'Supprimer ce fond d\'écran ?',

    // Effects
    'fx.themeMode':          'Thème de texte',
    'fx.themeMode.sub':      'Auto : s\'adapte à la luminosité du fond d\'écran',
    'fx.themeMode.auto':     'Auto',
    'fx.themeMode.dark':     'Sombre',
    'fx.themeMode.light':    'Clair',
    'fx.blur.bg':            'Flou du fond',
    'fx.blur.bg.sub':        'Quantité de flou sur l\'image de fond',
    'fx.blur.panel':         'Flou des panneaux',
    'fx.blur.panel.sub':     'Flou verre des panneaux et menus',
    'fx.blur.focus':         'Flou du focus',
    'fx.blur.focus.sub':     'Flou lors de l\'ouverture d\'un panneau',
    'fx.rain':               'Effet pluie',
    'fx.rain.sub':           'Animation de pluie sur la page d\'accueil',
    'fx.dust':               'Effet poussière',
    'fx.dust.sub':           'Particules zen qui dérivent lentement sur l\'écran',

    // Bookmarklet
    'bml.desc':              'Glissez le lien ci-dessous dans votre barre de favoris pour sauvegarder des pages en un clic.',
    'bml.generate':          'Générer un bookmarklet',
    'bml.copy':              'Copier le lien',
    'bml.copied':            'Copié ✓',
    'bml.revoke':            'Révoquer',
    'bml.revoke.confirm':    'Révoquer le bookmarklet ? Le lien actuel ne fonctionnera plus.',
    'bml.revoked':           'Révoqué. Générez-en un nouveau.',
    'bml.hint':              'Cliquez sur « Générer » pour créer un bookmarklet.',

    // Audit
    'audit.none':            'Aucune entrée.',

    // Stat labels
    'stat.bookmarks':        'Marque-pages',
    'stat.wallpapers':       'Fonds d\'écran',
    'stat.sessions':         'Sessions',
    'stat.member_since':     'Membre depuis',

    // Search
    'search.placeholder':    'Rechercher…',

    // Common
    'btn.save':              'Enregistrer',
    'btn.cancel':            'Annuler',
    'btn.delete':            'Supprimer',
    'btn.close':             'Fermer',
    'loading':               'Chargement…',
    'error':                 'Erreur',
    'error.network':         'Erreur réseau',
    'current':               'actuelle',
  },

  en: {
    // Clock
    days:   ['Sunday','Monday','Tuesday','Wednesday','Thursday','Friday','Saturday'],
    months: ['January','February','March','April','May','June','July','August','September','October','November','December'],

    // Login page
    'login.email':              'Email',
    'login.password':           'Password',
    'login.totp':               'TOTP Code',
    'login.submit':             'Sign in',
    'login.forgot':             'Forgot password?',
    'login.register':           'No account yet?',
    'login.sso':                'Sign in with',
    'login.or':                 'or',
    'forgot.back':              'Back',
    'forgot.hint':              'Enter your email and we\'ll send a reset link.',
    'forgot.submit':            'Send reset link',
    'forgot.sent':              'If the address exists, an email has been sent.',
    'register.username':        'Username',
    'register.email':           'Email',
    'register.password':        'Password',
    'register.password.hint':   'min. 12 characters',
    'register.hint':            'Enter your details — you\'ll receive a setup link by email.',
    'register.submit':          'Request account',
    'register.sent':            'Link sent! Check your inbox.',
    'register.back':            'Back',
    'setup.title':              'Account setup',
    'setup.totp.hint':          'Scan with your authenticator app',
    'setup.next':               'Continue',
    'setup.scan':               'Validate & continue',
    'setup.finish':             'Create account',
    'setup.invalid':            'This link is invalid or has expired.',
    'setup.password.hint':      'Choose a strong password (min. 12 characters)',
    'setup.password.confirm':   'Confirm password',
    'setup.password.mismatch':  'Passwords do not match.',
    'setup.open.app':           'Open in authenticator app',
    'btn.copy':                 'Copy',
    'account.delete':           'Delete account',
    'account.delete.desc':      'Permanent deletion — all your data will be erased (GDPR).',
    'account.delete.confirm':   'Enter your password to confirm account deletion.',
    'account.delete.btn':       'Delete permanently',
    'admin.delete.warning':     'Permanent deletion — all data for this user will be erased (GDPR). This cannot be undone.',

    // Hub tiles
    'hub.bookmarks':         'Bookmarks',
    'hub.bookmarks.sub':     'Browse & search',
    'hub.account':           'Account',
    'hub.account.sub':       'Profile, security',
    'hub.theme':             'Theme',
    'hub.theme.sub':         'Wallpapers & effects',
    'hub.admin':             'Administration',
    'hub.admin.sub':         'Users, settings',
    'hub.logout':            'Sign out',
    'hub.logout.sub':        'Close session',
    'hub.setup':             '⚙ Setup required — SMTP not configured. Click to configure.',

    // Panel titles & tabs
    'panel.bookmarks':       'Bookmarks',
    'panel.account':         'My account',
    'panel.theme':           'Theme',
    'panel.admin':           'Administration',
    'tab.profile':           'Profile',
    'tab.security':          'Security',
    'tab.engine':            'Search',
    'tab.sessions':          'Sessions',
    'tab.wallpapers':        'Wallpapers',
    'tab.effects':           'Effects',
    'tab.bookmarklet':       'Bookmarklet',
    'tab.stats':             'Statistics',
    'tab.users':             'Users',
    'tab.invitations':       'Invitations',
    'tab.settings':          'Settings',
    'tab.audit':             'Audit log',

    // Bookmarks panel
    'bm.add':                '+ Add',
    'bm.import':             'Import',
    'bm.export':             'Export',
    'bm.search.placeholder': 'Filter bookmarks…',
    'bm.all.folders':        'All folders',
    'bm.modal.add':          'Add bookmark',
    'bm.modal.edit':         'Edit bookmark',
    'bm.modal.title':        'Title',
    'bm.modal.title.ph':     'Bookmark title',
    'bm.modal.folder':       'Folder',
    'bm.modal.folder.ph':    'Dev/Go  (leave empty for none)',
    'bm.modal.tags':         'Tags',
    'bm.modal.tags.ph':      'homelab, selfhosted  (comma-separated)',
    'bm.empty':              'No bookmarks yet. Use + Add or Import.',
    'bm.no.folder':          'Uncategorised',
    'bm.url.required':       'URL and title are required.',
    'bm.confirm.delete':     'Delete this bookmark?',
    'bm.prev':               '← Previous',
    'bm.next':               'Next →',
    'bm.import.ok':          'Import done',
    'bm.import.added':       'added',
    'bm.import.skipped':     'skipped',

    // Profile tab
    'profile.stats':         'Statistics',
    'profile.title':         'Profile',
    'profile.username':      'Username',
    'profile.email':         'Email',
    'profile.save':          'Save',
    'profile.locale':        'Interface language',
    'profile.locale.fr':     'Français',
    'profile.locale.en':     'English',

    // Security tab
    'security.password':          'Change password',
    'security.current':           'Current password',
    'security.new':               'New password',
    'security.change':            'Update',
    'security.changed':           'Password updated',
    'security.totp':              'Two-factor authentication',
    'security.totp.note':         'TOTP adds a 6-digit code at login (Google Authenticator, etc.)',
    'security.totp.configure':    'Configure / Reset TOTP',
    'security.totp.enable':       'Enable 2FA',
    'security.totp.disable':      'Disable TOTP',
    'security.totp.code':         'Verification code',
    'security.totp.confirm':      'Confirm',
    'security.totp.confirm.prompt': 'Confirm with your current password:',

    // Sessions tab
    'sessions.title':        'Active sessions',
    'sessions.revoke':       'Revoke',
    'sessions.revoke_all':   'Revoke all sessions',
    'sessions.current':      '(current session)',
    'sessions.bookmarklet':  'Bookmarklet',
    'sessions.none':         'No sessions.',
    'sessions.revoke.all.confirm': 'Revoke all sessions (you will be signed out)?',

    // Section titles (JS-built panels)
    'section.engine':        'Search engine',
    'section.wallpapers':    'Wallpapers',
    'section.bookmarklet':   'Bookmarklet',
    'section.effects':       'Visual effects',
    'section.registration':  'Open registration',
    'admin.reg.enable':      'Disabled — enable',
    'admin.reg.disable':     'Enabled — disable',
    'admin.reg.desc':        'Allow public registration without an invitation',

    // Admin
    'admin.stat.users':      'Users',
    'admin.stat.active':     'Active',
    'admin.stat.bookmarks':  'Bookmarks',
    'admin.stat.wallpapers': 'Wallpapers',
    'admin.stat.db':         'Database',
    'admin.user.create':     'Create user',
    'admin.user.create.btn': 'Create',
    'admin.user.temp.password': 'Temporary password',
    'admin.user.suspend':    'Suspend',
    'admin.user.activate':   'Activate',
    'admin.user.delete':     'Delete',
    'admin.user.you':        '(you)',
    'admin.user.stats':      '… bookmarks · … wallpapers · … sessions',
    'admin.user.delete.confirm': 'Delete user {u}? This is irreversible.',
    'admin.inv.none':        'No invitations.',
    'admin.inv.used':        'used',
    'admin.inv.expired':     'expired',
    'admin.inv.pending':     'pending',
    'admin.inv.resend':      'Resend',
    'admin.inv.revoke':      'Revoke',
    'admin.inv.expires':     'expires',
    'admin.inv.invite':      'Invite',
    'admin.menu.label':      'Menu',
    'admin.menu.desc':       'The bang that opens the full-screen hub. Type it in the search bar.',
    'admin.menu.locked':     'Configured via compose (CAIRN_MENU_BANG) — not editable here.',
    'admin.saved':           'Saved.',
    'admin.section.system':  'System',
    'admin.section.infra':   'Infrastructure (compose · read-only)',
    'admin.section.smtp':    'Email (SMTP)',
    'admin.section.sso':     'SSO (OpenID Connect)',
    'admin.sso.desc':        'Connect an OIDC provider (Authentik, Keycloak…). The SSO button will appear on the login page.',
    'admin.upload.limit':      'Upload limit',
    'admin.upload.limit.save': 'Save limit',
    'admin.upload.limit.reset': 'Reset to global default',
    'admin.upload.global':     'global limit',
    'admin.quota':             'Storage quota',
    'admin.quota.save':        'Save quota',
    'admin.quota.reset':       'Reset to global default',
    'admin.limit.file':        'File (MB)',
    'admin.limit.storage':     'Storage (MB)',
    'admin.pending.title':     'Pending registration requests',
    'admin.pending.none':      'No pending requests.',
    'admin.pending.revoke':    'Revoke',
    'admin.pending.completed': 'completed',
    'admin.sys.addr':        'Listen address',
    'admin.sys.env':         'Environment',
    'admin.sys.base_url':    'Public URL',
    'admin.sys.db_path':     'Database',
    'admin.sys.media_path':  'Media directory',
    'admin.sys.max_upload':  'Max upload size',
    'admin.sys.storage_quota': 'Default storage quota',
    'admin.sys.trusted_proxy': 'Trusted proxy',
    'admin.sys.session_secret': 'Session secret',
    'admin.sys.version':     'Version',
    'admin.sys.set':         'set',
    'admin.sys.not_set':     '⚠ not set',
    'admin.sys.yes':         'yes',
    'admin.sys.no':          'no',

    // Engine
    'engine.custom':         'Custom',
    'engine.custom.prompt':  'Engine URL (must end with =):',

    // Wallpapers
    'wp.upload.label':       '+ Click or drag to upload a wallpaper',
    'wp.none':               'No wallpapers.',
    'wp.pin':                'Pin',
    'wp.unpin':              'Unpin',
    'wp.delete.confirm':     'Delete this wallpaper?',

    // Effects
    'fx.themeMode':          'Text theme',
    'fx.themeMode.sub':      'Auto: adapts to wallpaper brightness',
    'fx.themeMode.auto':     'Auto',
    'fx.themeMode.dark':     'Dark',
    'fx.themeMode.light':    'Light',
    'fx.blur.bg':            'Background blur',
    'fx.blur.bg.sub':        'Amount of blur applied to the background image',
    'fx.blur.panel':         'Panel blur',
    'fx.blur.panel.sub':     'Glass blur applied to panels and menus',
    'fx.blur.focus':         'Focus blur',
    'fx.blur.focus.sub':     'Background blur when a panel is open',
    'fx.rain':               'Rain effect',
    'fx.rain.sub':           'Rain animation on the home page',
    'fx.dust':               'Dust effect',
    'fx.dust.sub':           'Zen particles drifting slowly across the screen',

    // Bookmarklet
    'bml.desc':              'Drag the link below into your bookmarks bar to save pages in one click.',
    'bml.generate':          'Generate a bookmarklet',
    'bml.copy':              'Copy link',
    'bml.copied':            'Copied ✓',
    'bml.revoke':            'Revoke',
    'bml.revoke.confirm':    'Revoke the bookmarklet? The current link will stop working.',
    'bml.revoked':           'Revoked. Generate a new one.',
    'bml.hint':              'Click "Generate" to create a bookmarklet.',

    // Audit
    'audit.none':            'No entries.',

    // Stat labels
    'stat.bookmarks':        'Bookmarks',
    'stat.wallpapers':       'Wallpapers',
    'stat.sessions':         'Sessions',
    'stat.member_since':     'Member since',

    // Search
    'search.placeholder':    'Search…',

    // Common
    'btn.save':              'Save',
    'btn.cancel':            'Cancel',
    'btn.delete':            'Delete',
    'btn.close':             'Close',
    'loading':               'Loading…',
    'error':                 'Error',
    'error.network':         'Network error',
    'current':               'current',
  },
};

// Current locale — updated from the user profile at boot.
let _locale = 'en';

function t(key) {
  const dict = TRANSLATIONS[_locale] || TRANSLATIONS.fr;
  return dict[key] ?? TRANSLATIONS.fr[key] ?? key;
}

// Apply all data-i18n attributes in the document.
function applyLocale(locale) {
  _locale = locale || 'fr';
  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.getAttribute('data-i18n');
    const attr = el.getAttribute('data-i18n-attr');
    if (attr) {
      el.setAttribute(attr, t(key));
    } else {
      el.textContent = t(key);
    }
  });
}

/* ─── State ──────────────────────────────────────────────────────────────── */
const S = {
  user:        null,
  bookmarks:   [],
  bmTotal:     0,
  bmOffset:    0,
  bmLimit:     50,
  bmFilter:    '',
  bmFolder:    '',
  wallpapers:  [],
  sessions:    [],
  editingBmId: null,
  menuBang:    '!menu',
};

/* ─── API client ─────────────────────────────────────────────────────────── */
async function api(method, path, body) {
  const opts = {
    method,
    credentials: 'same-origin',
    headers: body ? { 'Content-Type': 'application/json' } : {},
    body: body ? JSON.stringify(body) : undefined,
  };
  const r = await fetch('/api' + path, opts);
  if (r.status === 204) return null;
  const json = await r.json().catch(() => ({}));
  if (!r.ok) {
    const err = new Error(json.error || r.statusText || 'Network error');
    err.code   = json.code;
    err.status = r.status;
    throw err;
  }
  return json;
}

const GET  = path       => api('GET',    path);
const POST = (path, b)  => api('POST',   path, b);
const PUT  = (path, b)  => api('PUT',    path, b);
const DEL  = (path, b)  => api('DELETE', path, b);

/* ─── DOM helpers ────────────────────────────────────────────────────────── */
const $  = id => document.getElementById(id);
const el = (tag, cls, text) => {
  const e = document.createElement(tag);
  if (cls)  e.className   = cls;
  if (text) e.textContent = text;
  return e;
};

function show(id)  { $(id).classList.remove('hidden'); }
function hide(id)  { $(id).classList.add('hidden'); }
function toggle(id, on) { $(id).classList.toggle('hidden', !on); }

function setError(id, msg) { $(id).textContent = msg || ''; }

/* ─── Rain canvas ────────────────────────────────────────────────────────── */
function initRain() {
  const canvas = $('rain-canvas');
  const ctx    = canvas.getContext('2d');
  let drops    = [];

  function resize() {
    canvas.width  = window.innerWidth;
    canvas.height = window.innerHeight;
    const count = Math.floor(canvas.width / 14);
    drops = Array.from({ length: count }, () => ({
      x:       Math.random() * canvas.width,
      y:       Math.random() * canvas.height,
      speed:   7 + Math.random() * 10,
      length:  12 + Math.random() * 18,
      opacity: 0.04 + Math.random() * 0.12,
    }));
  }

  function frame() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    for (const d of drops) {
      ctx.beginPath();
      ctx.strokeStyle = `rgba(190,215,255,${d.opacity})`;
      ctx.lineWidth   = 1;
      ctx.moveTo(d.x, d.y);
      ctx.lineTo(d.x - 0.8, d.y + d.length);
      ctx.stroke();
      d.y += d.speed * 0.016 * 60; // ~60fps target
      if (d.y - d.length > canvas.height) {
        d.y = -d.length;
        d.x = Math.random() * canvas.width;
      }
    }
    requestAnimationFrame(frame);
  }

  window.addEventListener('resize', resize);
  resize();
  requestAnimationFrame(frame);
}

/* ─── Dust canvas — slow zen particles drifting across the screen ────────── */
function initDust() {
  const canvas = $('dust-canvas');
  if (!canvas) return;
  const ctx  = canvas.getContext('2d');
  let motes  = [];

  function resize() {
    canvas.width  = window.innerWidth;
    canvas.height = window.innerHeight;
    const count = Math.floor((canvas.width * canvas.height) / 26000);
    motes = Array.from({ length: count }, () => spawn(true));
  }

  function spawn(anywhere) {
    return {
      x:     anywhere ? Math.random() * canvas.width : -8,
      y:     Math.random() * canvas.height,
      r:     0.6 + Math.random() * 1.8,           // tiny soft specks
      vx:    0.08 + Math.random() * 0.25,          // slow lateral drift
      vy:    -0.05 + Math.random() * 0.1,          // barely rises/sinks
      phase: Math.random() * Math.PI * 2,          // twinkle offset
      tw:    0.4 + Math.random() * 0.8,            // twinkle speed
      base:  0.05 + Math.random() * 0.16,          // base opacity
    };
  }

  function frame(now) {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    const tSec = now / 1000;
    for (const m of motes) {
      // gentle sine sway + slow twinkle
      const sway    = Math.sin(tSec * 0.3 + m.phase) * 0.15;
      const opacity = m.base * (0.6 + 0.4 * Math.sin(tSec * m.tw + m.phase));
      ctx.beginPath();
      ctx.fillStyle = `rgba(255,250,235,${Math.max(opacity, 0.015)})`;
      ctx.arc(m.x, m.y, m.r, 0, Math.PI * 2);
      ctx.fill();
      m.x += m.vx;
      m.y += m.vy + sway;
      if (m.x - m.r > canvas.width || m.y < -10 || m.y > canvas.height + 10) {
        Object.assign(m, spawn(false));
      }
    }
    requestAnimationFrame(frame);
  }

  window.addEventListener('resize', resize);
  resize();
  requestAnimationFrame(frame);
}

/* ─── Clock ──────────────────────────────────────────────────────────────── */
function startClock() {
  const DAYS   = () => t('days');
  const MONTHS = () => t('months');
  function isoWeek(d) {
    const date = new Date(Date.UTC(d.getFullYear(), d.getMonth(), d.getDate()));
    date.setUTCDate(date.getUTCDate() + 4 - (date.getUTCDay() || 7));
    const y0 = new Date(Date.UTC(date.getUTCFullYear(), 0, 1));
    return Math.ceil((((date - y0) / 86400000) + 1) / 7);
  }

  function tick() {
    const now = new Date();
    const h = String(now.getHours()).padStart(2, '0');
    const m = String(now.getMinutes()).padStart(2, '0');
    $('clock').textContent    = `${h}:${m}`;
    $('date-line').textContent =
      `${DAYS()[now.getDay()]} ${now.getDate()} ${MONTHS()[now.getMonth()]} ${now.getFullYear()} · S${isoWeek(now)}`;
  }

  tick();
  setInterval(tick, 1000);
}

/* ─── Wallpaper ──────────────────────────────────────────────────────────── */
async function loadWallpaper() {
  const wps = S.wallpapers;
  if (!wps.length) return; // gradient fallback stays

  const pinned = wps.filter(w => w.is_pinned);
  const pool   = pinned.length ? pinned : wps;
  const wp     = pool[Math.floor(Math.random() * pool.length)];
  const url    = `/media/${S.user.id}/${wp.filename}`;
  const ext    = wp.filename.split('.').pop().toLowerCase();
  const isVid  = ['mp4','webm'].includes(ext);

  if (isVid) {
    const v   = $('bg-video');
    v.src     = url;
    v.classList.remove('bg-hidden');
    $('bg-gradient').classList.add('bg-hidden');
    v.addEventListener('loadeddata', () => sampleLuminance(v), { once: true });
    // Video brightness drifts during playback — re-sample so the theme follows.
    if (_lumInterval) clearInterval(_lumInterval);
    _lumInterval = setInterval(() => { if (!v.paused) sampleLuminance(v); }, 10000);
  } else {
    if (_lumInterval) { clearInterval(_lumInterval); _lumInterval = null; }
    const img = $('bg-image');
    img.src   = url;
    img.classList.remove('bg-hidden');
    $('bg-gradient').classList.add('bg-hidden');
    img.addEventListener('load', () => sampleLuminance(img), { once: true });
  }
}

let _lumInterval = null;
let _bgMedia = null; // current background element, for re-sampling on theme-mode change

function themeMode() {
  const m = loadThemePrefs().themeMode;
  return m === 'dark' || m === 'light' ? m : 'auto';
}

function resampleTheme() {
  const mode = themeMode();
  if (mode !== 'auto') {
    document.documentElement.dataset.theme = mode;
    return;
  }
  if (_bgMedia) sampleLuminance(_bgMedia);
  else document.documentElement.dataset.theme = 'dark'; // gradient fallback is dark
}

// Resizing changes the visible cover crop → the sampled band moves with it.
let _resampleTimer = null;
window.addEventListener('resize', () => {
  clearTimeout(_resampleTimer);
  _resampleTimer = setTimeout(resampleTheme, 300);
});

function sampleLuminance(media) {
  _bgMedia = media;
  if (themeMode() !== 'auto') return;
  // The background uses object-fit: cover, so what's on screen is a crop of
  // the source — a fixed center crop of the source can land on a region the
  // text never overlaps. Reproduce the cover mapping, then sample the wide
  // band where clock, date and search actually sit.
  // Some browsers taint the canvas even for same-origin video — wrap carefully.
  try {
    const sw = media.videoWidth || media.naturalWidth || media.offsetWidth;
    const sh = media.videoHeight || media.naturalHeight || media.offsetHeight;
    if (!sw || !sh) return;
    const vw = window.innerWidth, vh = window.innerHeight;
    // object-fit: cover → source rect actually visible on screen
    const scale = Math.max(vw / sw, vh / sh);
    const visW = vw / scale, visH = vh / scale;
    const offX = (sw - visW) / 2, offY = (sh - visH) / 2;
    // Text band: middle 70% of width, from above the clock to below the search bar
    const bx = offX + visW * 0.15, bw = visW * 0.70;
    const by = offY + visH * 0.18, bh = visH * 0.62;
    const W = 96, H = 60;
    const c = document.createElement('canvas');
    c.width = W; c.height = H;
    const ctx = c.getContext('2d', { willReadFrequently: true });
    ctx.drawImage(media, bx, by, bw, bh, 0, 0, W, H);
    const d = ctx.getImageData(0, 0, W, H).data;
    let sum = 0, n = 0;
    for (let i = 0; i < d.length; i += 8) {
      sum += 0.299 * d[i] + 0.587 * d[i+1] + 0.114 * d[i+2];
      n++;
    }
    const lum = sum / n;
    document.documentElement.dataset.theme = lum > 140 ? 'light' : 'dark';
  } catch {
    // Tainted canvas (video DRM or browser restriction) — keep current theme
  }
}

/* ─── Search ─────────────────────────────────────────────────────────────── */
const ENGINES = {
  duckduckgo: 'https://duckduckgo.com/?q=',
  google:     'https://www.google.com/search?q=',
  brave:      'https://search.brave.com/search?q=',
  bing:       'https://www.bing.com/search?q=',
  kagi:       'https://kagi.com/search?q=',
};

function engineURL() {
  const u = S.user;
  if (!u) return ENGINES.duckduckgo;
  if (u.search_engine === 'custom' && u.search_engine_url) return u.search_engine_url;
  return ENGINES[u.search_engine] || ENGINES.duckduckgo;
}

const BANGS = [
  { bang: '!g',     label: 'Google',          url: 'https://www.google.com/search?q=' },
  { bang: '!yt',    label: 'YouTube',          url: 'https://www.youtube.com/results?search_query=' },
  { bang: '!gh',    label: 'GitHub',           url: 'https://github.com/search?q=' },
  { bang: '!hub',   label: 'Docker Hub',       url: 'https://hub.docker.com/search?q=' },
  { bang: '!ddg',   label: 'DuckDuckGo',       url: 'https://duckduckgo.com/?q=' },
  { bang: '!b',     label: 'Bing',             url: 'https://www.bing.com/search?q=' },
  { bang: '!br',    label: 'Brave Search',     url: 'https://search.brave.com/search?q=' },
  { bang: '!kagi',  label: 'Kagi',             url: 'https://kagi.com/search?q=' },
  { bang: '!az',    label: 'Amazon',           url: 'https://www.amazon.com/s?k=' },
  { bang: '!afr',   label: 'Amazon.fr',        url: 'https://www.amazon.fr/s?k=' },
  { bang: '!aze',   label: 'Amazon.es',        url: 'https://www.amazon.es/s?k=' },
  { bang: '!w',     label: 'Wikipedia',        url: 'https://en.wikipedia.org/w/index.php?search=' },
  { bang: '!wfr',   label: 'Wikipédia (FR)',   url: 'https://fr.wikipedia.org/w/index.php?search=' },
  { bang: '!maps',  label: 'Google Maps',      url: 'https://www.google.com/maps/search/' },
  { bang: '!img',   label: 'Google Images',    url: 'https://www.google.com/search?tbm=isch&q=' },
  { bang: '!tw',    label: 'X / Twitter',      url: 'https://twitter.com/search?q=' },
  { bang: '!rd',    label: 'Reddit',           url: 'https://www.reddit.com/search/?q=' },
  { bang: '!so',    label: 'Stack Overflow',   url: 'https://stackoverflow.com/search?q=' },
  { bang: '!mdn',   label: 'MDN',              url: 'https://developer.mozilla.org/search?q=' },
  { bang: '!npm',   label: 'npm',              url: 'https://www.npmjs.com/search?q=' },
  { bang: '!pkg',   label: 'pkg.go.dev',       url: 'https://pkg.go.dev/search?q=' },
  { bang: '!pypi',  label: 'PyPI',             url: 'https://pypi.org/search/?q=' },
  { bang: '!cr',    label: 'Crates.io',        url: 'https://crates.io/search?q=' },
  { bang: '!tf',    label: 'Terraform Registry', url: 'https://registry.terraform.io/search/providers?q=' },
  { bang: '!ia',    label: 'Internet Archive', url: 'https://archive.org/search?query=' },
  { bang: '!li',    label: 'LinkedIn',         url: 'https://www.linkedin.com/search/results/all/?keywords=' },
  { bang: '!insta', label: 'Instagram',        url: 'https://www.instagram.com/explore/tags/' },
  { bang: '!pin',   label: 'Pinterest',        url: 'https://www.pinterest.com/search/pins/?q=' },
  { bang: '!wp',    label: 'WordPress',        url: 'https://wordpress.org/search/' },
  { bang: '!leo',   label: 'Leo (dict)',        url: 'https://dict.leo.org/englisch-deutsch/' },
  { bang: '!tr',    label: 'DeepL',            url: 'https://www.deepl.com/translator#auto/auto/' },
  { bang: '!bm',    label: 'Mes marque-pages', url: null }, // handled specially
];

function initSearchSuggestions() {
  const input = $('search-input');
  const box   = $('search-suggestions');
  if (!input || !box) return;

  let debounce = null;
  let activeIdx = -1;
  let items = [];

  function hideSuggestions() {
    box.classList.remove('visible');
    box.innerHTML = '';
    activeIdx = -1;
    items = [];
  }

  function setActive(idx) {
    items.forEach((it, i) => it.classList.toggle('active', i === idx));
    activeIdx = idx;
  }

  function buildBangRow(bang, rest) {
    const row = el('div', 'sug-item');
    row.appendChild(el('span', 'sug-title', bang.bang + (rest ? ' ' + rest : '')));
    row.appendChild(el('span', 'sug-url', bang.label));
    row.addEventListener('mousedown', e => {
      e.preventDefault();
      if (!bang.url) { openHub(); }
      else if (rest) open(bang.url + encodeURIComponent(rest), '_blank');
      else { input.value = bang.bang + ' '; hideSuggestions(); input.focus(); return; }
      hideSuggestions();
      input.value = '';
    });
    return row;
  }

  function showBangSuggestions(q) {
    const m = q.match(/^!(\S*)(?:\s(.*))?$/);
    if (!m) { hideSuggestions(); return; }
    const typed = ('!' + (m[1] || '')).toLowerCase();
    const rest  = (m[2] || '').trim();

    // Prepend the configurable menu bang if it matches the typed prefix.
    const menuBangEntry = S.menuBang && S.menuBang.toLowerCase().startsWith(typed)
      ? [{ bang: S.menuBang, label: 'Menu Cairn', url: null }]
      : [];

    const matched = [...menuBangEntry, ...BANGS.filter(b => b.bang.startsWith(typed))].slice(0, 6);
    if (!matched.length) { hideSuggestions(); return; }

    box.innerHTML = '';
    items = [];
    activeIdx = -1;
    for (const bang of matched) {
      const row = buildBangRow(bang, rest);
      box.appendChild(row);
      items.push(row);
    }
    box.classList.add('visible');
  }

  async function showBookmarkSuggestions(q) {
    try {
      const params = new URLSearchParams({ search: q, limit: 6 });
      const data   = await GET(`/bookmarks?${params}`);
      const bms    = data.bookmarks || [];
      if (!bms.length) { hideSuggestions(); return; }

      box.innerHTML = '';
      items = [];
      activeIdx = -1;

      for (const bm of bms) {
        const row = el('div', 'sug-item');
        row.appendChild(el('span', 'sug-title', bm.title || bm.url));

        if (bm.tags && bm.tags.length) {
          const tagWrap = el('span', 'sug-tags');
          bm.tags.slice(0, 3).forEach(tag => tagWrap.appendChild(el('span', 'sug-tag', tag.name)));
          row.appendChild(tagWrap);
        }

        try {
          row.appendChild(el('span', 'sug-url', new URL(bm.url).hostname.replace(/^www\./, '')));
        } catch {}

        row.addEventListener('mousedown', e => {
          e.preventDefault();
          open(bm.url, '_blank');
          hideSuggestions();
          input.value = '';
        });

        box.appendChild(row);
        items.push(row);
      }

      box.classList.add('visible');
    } catch { hideSuggestions(); }
  }

  function onInput() {
    const q = input.value.trim();
    if (!q) { hideSuggestions(); return; }
    clearTimeout(debounce);
    if (q.startsWith('!')) {
      showBangSuggestions(q);
    } else {
      debounce = setTimeout(() => showBookmarkSuggestions(q), 180);
    }
  }

  input.addEventListener('input', onInput);

  input.addEventListener('keydown', e => {
    if (!box.classList.contains('visible')) return;
    if (e.key === 'ArrowDown')  { e.preventDefault(); setActive(Math.min(activeIdx + 1, items.length - 1)); }
    if (e.key === 'ArrowUp')    { e.preventDefault(); setActive(Math.max(activeIdx - 1, 0)); }
    if (e.key === 'Escape')     { hideSuggestions(); }
    if (e.key === 'Enter' && activeIdx >= 0) {
      e.preventDefault();
      items[activeIdx].dispatchEvent(new MouseEvent('mousedown'));
    }
  });

  input.addEventListener('blur', () => setTimeout(hideSuggestions, 150));
}

function handleSearch(e) {
  e.preventDefault();
  const q = $('search-input').value.trim();
  if (!q) return;

  // Menu bang (configurable) → open the full-page hub.
  if (q.toLowerCase() === S.menuBang.toLowerCase()) {
    openHub();
    $('search-input').value = '';
    return;
  }

  const m = q.match(/^!(\w+)\s*(.*)/s);
  if (m) {
    const bang = m[1].toLowerCase();
    const rest = m[2].trim();
    switch (bang) {
      case 'bm': case 'edit':
        openBookmarks();
        $('search-input').value = '';
        return;
      case 'g':
        open(`https://www.google.com/search?q=${encodeURIComponent(rest)}`);
        return;
      case 'yt':
        open(`https://www.youtube.com/results?search_query=${encodeURIComponent(rest)}`);
        return;
      case 'gh':
        open(`https://github.com/search?q=${encodeURIComponent(rest)}`);
        return;
      case 'hub':
        open(`https://hub.docker.com/search?q=${encodeURIComponent(rest)}`);
        return;
      default:
        // pass all other DDG bangs through
        open(`https://duckduckgo.com/?q=${encodeURIComponent(q)}`);
        return;
    }
  }
  open(engineURL() + encodeURIComponent(q));
}

function open(url) { window.location.href = url; }

/* ─── Hub (menu bang) ────────────────────────────────────────────────────── */
function openHub() {
  const isAdmin = S.user && S.user.role === 'admin';
  toggle('tile-admin', isAdmin);
  // Setup prompt: admins see it when a mandatory prerequisite (SMTP) is missing.
  toggle('hub-setup', isAdmin && S.user && S.user.smtp_configured === false);
  const g = $('hub-greeting');
  if (g && S.user) g.textContent = S.user.username;
  show('hub-overlay');
}
function closeHub() {
  const el = $('hub-overlay');
  el.classList.add('closing');
  setTimeout(() => { el.classList.remove('closing'); hide('hub-overlay'); }, 220);
}

// Close a sub-panel and step back to the hub menu (one level up).
function backToHub(panelId) {
  hide(panelId);
  openHub();
}

/* ─── Login ──────────────────────────────────────────────────────────────── */
async function handleLogin(e) {
  e.preventDefault();
  setError('login-error', '');
  const email    = $('login-email').value.trim();
  const password = $('login-password').value;
  const totpCode = $('login-totp').value.trim();

  try {
    await POST('/auth/login', { email, password, totp_code: totpCode || undefined });
    await boot();
  } catch (err) {
    if (err.code === 'TOTP_REQUIRED') {
      show('totp-group');
      $('login-totp').focus();
      setError('login-error', 'Code TOTP requis');
    } else {
      setError('login-error', err.message);
    }
  }
}

async function handleForgot(e) {
  e.preventDefault();
  setError('forgot-msg', '');
  const email = $('forgot-email').value.trim();
  try {
    await POST('/auth/forgot-password', { email });
    setError('forgot-msg', t('forgot.sent'));
  } catch {
    setError('forgot-msg', t('forgot.sent'));
  }
}

function showForgotForm() {
  hide('login-form');
  hide('register-form');
  hide('login-links');
  show('forgot-form');
}

function showRegisterForm() {
  hide('login-form');
  hide('forgot-form');
  hide('login-links');
  setError('reg-error', '');
  hide('reg-success');
  show('reg-inputs');
  show('register-form');
  $('reg-submit').disabled = true;
  $('reg-username').focus();
}

function showLoginForm() {
  hide('forgot-form');
  hide('register-form');
  show('login-links');
  show('login-form');
  setError('forgot-msg', '');
}

// Open-registration step 1: send username + email, receive a setup link by email.
async function handleRegister(e) {
  e.preventDefault();
  setError('reg-error', '');
  const username = $('reg-username').value.trim();
  const email    = $('reg-email').value.trim();
  const btn      = $('reg-submit');
  btn.disabled   = true;
  try {
    await POST('/auth/request-registration', { username, email });
    $('reg-username').value = '';
    $('reg-email').value    = '';
    $('reg-success-msg').textContent = t('register.sent');
    hide('reg-inputs');
    show('reg-success');
  } catch (err) {
    setError('reg-error', err.message);
    btn.disabled = false;
  }
}

/* ─── Logout ─────────────────────────────────────────────────────────────── */
async function logout() {
  try { await POST('/auth/logout'); } catch {}
  S.user = null;
  hide('view-home');
  hide('overlay-bookmarks');
  hide('overlay-settings');
  hide('overlay-admin');
  show('view-login');
  $('login-email').focus();
}

/* ─── Bookmark panel ─────────────────────────────────────────────────────── */
function openBookmarks() {
  show('overlay-bookmarks');
  loadBookmarks();
}

function closeBookmarks() {
  backToHub('overlay-bookmarks');
}

/* ─── Theme panel ────────────────────────────────────────────────────────── */
function openTheme() {
  renderThemeTab('wallpapers');
  show('overlay-theme');
}

function closeTheme() {
  backToHub('overlay-theme');
}

function renderThemeTab(tabName) {
  document.querySelectorAll('#theme-tabs .tab').forEach(t => {
    t.classList.toggle('active', t.dataset.tab === tabName);
  });
  const body = $('theme-body');
  body.innerHTML = '';
  if (tabName === 'wallpapers') body.appendChild(buildWallpapersTab());
  if (tabName === 'effects')    body.appendChild(buildEffectsTab());
}

function buildEffectsTab() {
  const frag = document.createDocumentFragment();
  const sec  = el('div', 'settings-section');
  sec.appendChild(el('div', 'settings-section-title', t('section.effects')));

  const prefs = loadThemePrefs();

  function makeSliderRow(label, sub, key, min, max, unit, defaultVal) {
    const row = el('div', 'effect-row');
    const labelWrap = el('div');
    labelWrap.appendChild(el('div', 'effect-label', label));
    if (sub) labelWrap.appendChild(el('div', 'effect-sub', sub));
    row.appendChild(labelWrap);

    const ctrl = el('div', 'effect-ctrl');
    const slider = el('input', 'blur-slider');
    slider.type = 'range'; slider.min = min; slider.max = max;
    slider.value = prefs[key] ?? defaultVal;
    const valLabel = el('span', 'blur-val', slider.value + unit);

    slider.oninput = () => {
      valLabel.textContent = slider.value + unit;
      const newPrefs = loadThemePrefs();
      newPrefs[key] = parseInt(slider.value, 10);
      saveThemePrefs(newPrefs);
      applyThemePrefs(newPrefs);
    };

    ctrl.append(slider, valLabel);
    row.appendChild(ctrl);
    return row;
  }

  function makeToggleRow(label, sub, key, defaultVal) {
    const row = el('div', 'effect-row');
    const labelWrap = el('div');
    labelWrap.appendChild(el('div', 'effect-label', label));
    if (sub) labelWrap.appendChild(el('div', 'effect-sub', sub));
    row.appendChild(labelWrap);

    const toggle = el('label', 'toggle-switch');
    const inp = el('input'); inp.type = 'checkbox';
    inp.checked = prefs[key] ?? defaultVal;
    const track = el('span', 'toggle-track');
    toggle.append(inp, track);

    inp.onchange = () => {
      const newPrefs = loadThemePrefs();
      newPrefs[key] = inp.checked;
      saveThemePrefs(newPrefs);
      applyThemePrefs(newPrefs);
    };

    row.appendChild(toggle);
    return row;
  }

  function makeThemeModeRow() {
    const row = el('div', 'effect-row');
    const labelWrap = el('div');
    labelWrap.appendChild(el('div', 'effect-label', t('fx.themeMode')));
    labelWrap.appendChild(el('div', 'effect-sub', t('fx.themeMode.sub')));
    row.appendChild(labelWrap);

    const seg = el('div', 'segmented');
    const modes = [
      ['auto',  t('fx.themeMode.auto')],
      ['dark',  t('fx.themeMode.dark')],
      ['light', t('fx.themeMode.light')],
    ];
    const current = prefs.themeMode === 'dark' || prefs.themeMode === 'light' ? prefs.themeMode : 'auto';
    for (const [mode, label] of modes) {
      const btn = el('button', mode === current ? 'active' : '', label);
      btn.type = 'button';
      btn.onclick = () => {
        seg.querySelectorAll('button').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        const newPrefs = loadThemePrefs();
        newPrefs.themeMode = mode;
        saveThemePrefs(newPrefs);
        applyThemePrefs(newPrefs);
      };
      seg.appendChild(btn);
    }
    row.appendChild(seg);
    return row;
  }

  sec.appendChild(makeThemeModeRow());
  sec.appendChild(makeSliderRow(t('fx.blur.bg'), t('fx.blur.bg.sub'), 'blurBg', 0, 40, 'px', 0));
  sec.appendChild(makeSliderRow(t('fx.blur.panel'), t('fx.blur.panel.sub'), 'blurPanel', 10, 60, 'px', 40));
  sec.appendChild(makeSliderRow(t('fx.blur.focus'), t('fx.blur.focus.sub'), 'blurFocus', 0, 30, 'px', 14));
  sec.appendChild(makeToggleRow(t('fx.rain'), t('fx.rain.sub'), 'rain', true));
  sec.appendChild(makeToggleRow(t('fx.dust'), t('fx.dust.sub'), 'dust', false));

  frag.appendChild(sec);
  return frag;
}

function loadThemePrefs() {
  try { return JSON.parse(localStorage.getItem('cairn_theme') || '{}'); } catch { return {}; }
}

function saveThemePrefs(p) {
  localStorage.setItem('cairn_theme', JSON.stringify(p));
}

function applyThemePrefs(p) {
  const root = document.documentElement.style;
  if (p.blurBg    !== undefined) root.setProperty('--blur-bg',    p.blurBg    + 'px');
  if (p.blurPanel !== undefined) root.setProperty('--blur-panel', p.blurPanel + 'px');
  if (p.blurFocus !== undefined) root.setProperty('--blur-focus', p.blurFocus + 'px');

  const canvas = $('rain-canvas');
  if (canvas) canvas.style.display = (p.rain === false) ? 'none' : '';
  const dust = $('dust-canvas');
  if (dust) dust.style.display = (p.dust === true) ? '' : 'none'; // opt-in

  resampleTheme();
}

async function loadBookmarks() {
  const params = new URLSearchParams({
    offset: S.bmOffset,
    limit:  S.bmLimit,
  });
  if (S.bmFilter) params.set('search', S.bmFilter);
  if (S.bmFolder) params.set('folder', S.bmFolder);

  try {
    const data = await GET(`/bookmarks?${params}`);
    S.bookmarks = data.bookmarks || data.items || data || [];
    S.bmTotal   = data.total ?? S.bookmarks.length;
    renderBookmarks();
  } catch (err) {
    $('bm-list').textContent = t('error') + ': ' + err.message;
  }
}

function renderBookmarks() {
  const list = $('bm-list');
  list.innerHTML = '';

  if (!S.bookmarks.length) {
    const empty = el('p', 'bm-empty', t('bm.empty'));
    list.appendChild(empty);
    return;
  }

  // Group by folder
  const byFolder = new Map();
  const noFolder = [];
  for (const bm of S.bookmarks) {
    if (bm.folder) {
      if (!byFolder.has(bm.folder)) byFolder.set(bm.folder, []);
      byFolder.get(bm.folder).push(bm);
    } else {
      noFolder.push(bm);
    }
  }

  // Render folders
  for (const [folder, bms] of [...byFolder.entries()].sort()) {
    const section = el('div', 'bm-section');
    section.appendChild(el('div', 'bm-section-name', folder));
    const count = bms.length;
    bms.forEach((bm, i) => {
      section.appendChild(makeBmItem(bm, i < count - 1 ? '├' : '└'));
    });
    list.appendChild(section);
  }

  // Render unfoldered
  if (noFolder.length) {
    const section = el('div', 'bm-section');
    section.appendChild(el('div', 'bm-section-name', t('bm.no.folder')));
    noFolder.forEach(bm => section.appendChild(makeBmItem(bm, '·')));
    list.appendChild(section);
  }

  // Pagination
  if (S.bmTotal > S.bmLimit) {
    const pag = el('div', 'pagination');
    const prev = el('button', 'page-btn', t('bm.prev'));
    prev.disabled = S.bmOffset === 0;
    prev.onclick  = () => { S.bmOffset = Math.max(0, S.bmOffset - S.bmLimit); loadBookmarks(); };

    const info = el('span', 'page-info',
      `${S.bmOffset + 1}–${Math.min(S.bmOffset + S.bmLimit, S.bmTotal)} / ${S.bmTotal}`);

    const next = el('button', 'page-btn', t('bm.next'));
    next.disabled = S.bmOffset + S.bmLimit >= S.bmTotal;
    next.onclick  = () => { S.bmOffset += S.bmLimit; loadBookmarks(); };

    pag.append(prev, info, next);
    list.appendChild(pag);
  }

  // Populate folder filter
  const sel = $('bm-folder-filter');
  const current = sel.value;
  sel.innerHTML = `<option value="">${t('bm.all.folders')}</option>`;
  for (const f of [...byFolder.keys()].sort()) {
    const opt = document.createElement('option');
    opt.value       = f;
    opt.textContent = f;
    if (f === current) opt.selected = true;
    sel.appendChild(opt);
  }
}

function makeBmItem(bm, glyph) {
  const wrap = el('div', 'bm-item');

  const g = el('span', 'bm-tree-glyph', glyph);

  const link = el('a', 'bm-link', bm.title || bm.url);
  link.href   = bm.url;
  link.target = '_blank';
  link.rel    = 'noopener noreferrer';

  const urlSpan = el('span', 'bm-url', new URL(bm.url).hostname);

  const actions = el('div', 'bm-actions');

  const editBtn = el('button', 'icon-btn', '✎');
  editBtn.title = 'Modifier';
  editBtn.onclick = () => openEditBookmark(bm);

  const delBtn = el('button', 'icon-btn danger', '✕');
  delBtn.title = t('btn.delete');
  delBtn.onclick = () => deleteBookmark(bm.id);

  actions.append(editBtn, delBtn);
  wrap.append(g, link, urlSpan, actions);

  // Tags row
  if (bm.tags && bm.tags.length) {
    const tagsRow = el('div', 'bm-tags-row');
    for (const tag of bm.tags) {
      tagsRow.appendChild(el('span', 'tag-chip', '#' + tag.name));
    }
    // Return a fragment with both rows
    const frag = document.createDocumentFragment();
    frag.append(wrap, tagsRow);
    return frag;
  }

  return wrap;
}

function openAddBookmark() {
  S.editingBmId = null;
  $('modal-bm-title').textContent   = t('bm.modal.add');
  $('modal-bm-url').value           = '';
  $('modal-bm-title-input').value   = '';
  $('modal-bm-folder').value        = '';
  $('modal-bm-tags').value          = '';
  setError('modal-bm-error', '');
  show('modal-bookmark');
  $('modal-bm-url').focus();
}

function openEditBookmark(bm) {
  S.editingBmId = bm.id;
  $('modal-bm-title').textContent   = t('bm.modal.edit');
  $('modal-bm-url').value           = bm.url;
  $('modal-bm-title-input').value   = bm.title;
  $('modal-bm-folder').value        = bm.folder || '';
  $('modal-bm-tags').value          = (bm.tags || []).map(t => t.name).join(', ');
  setError('modal-bm-error', '');
  show('modal-bookmark');
  $('modal-bm-title-input').focus();
}

async function saveBookmark() {
  setError('modal-bm-error', '');
  const url    = $('modal-bm-url').value.trim();
  const title  = $('modal-bm-title-input').value.trim();
  const folder = $('modal-bm-folder').value.trim() || null;
  const tags   = $('modal-bm-tags').value.split(',').map(t => t.trim()).filter(Boolean);

  if (!url || !title) {
    setError('modal-bm-error', t('bm.url.required'));
    return;
  }

  try {
    if (S.editingBmId) {
      await PUT(`/bookmarks/${S.editingBmId}`, { url, title, folder, tags });
    } else {
      await POST('/bookmarks', { url, title, folder, tags });
    }
    hide('modal-bookmark');
    S.bmOffset = 0;
    loadBookmarks();
  } catch (err) {
    setError('modal-bm-error', err.message);
  }
}

async function deleteBookmark(id) {
  if (!confirm(t('bm.confirm.delete'))) return;
  try {
    await DEL(`/bookmarks/${id}`);
    loadBookmarks();
  } catch (err) {
    alert(err.message);
  }
}

async function exportBookmarks() {
  try {
    const r = await fetch('/api/bookmarks/export', { credentials: 'same-origin' });
    if (!r.ok) throw new Error(t('error.network'));
    const blob = await r.blob();
    const a    = document.createElement('a');
    a.href     = URL.createObjectURL(blob);
    a.download = 'bookmarks.html';
    a.click();
    URL.revokeObjectURL(a.href);
  } catch (err) {
    alert(err.message);
  }
}

async function importBookmarks(file) {
  if (!file) return;
  const form = new FormData();
  form.append('file', file);
  try {
    const r = await fetch('/api/bookmarks/import', {
      method: 'POST',
      credentials: 'same-origin',
      body: form,
    });
    const json = await r.json().catch(() => ({}));
    if (!r.ok) throw new Error(json.error || 'Import error');
    alert(`${t('bm.import.ok')}: ${json.imported ?? '?'} ${t('bm.import.added')}, ${json.skipped ?? '?'} ${t('bm.import.skipped')}.`);
    S.bmOffset = 0;
    loadBookmarks();
  } catch (err) {
    alert(err.message);
  }
}

/* ─── Settings panel ─────────────────────────────────────────────────────── */
function openSettings() {
  renderSettingsTab('profile');
  show('overlay-settings');
}

function closeSettings() {
  backToHub('overlay-settings');
}

function renderSettingsTab(tabName) {
  // Update tab highlights
  document.querySelectorAll('#settings-tabs .tab').forEach(t => {
    t.classList.toggle('active', t.dataset.tab === tabName);
  });

  const body = $('settings-body');
  body.innerHTML = '';

  switch (tabName) {
    case 'profile':     body.appendChild(buildProfileTab());     break;
    case 'security':    body.appendChild(buildSecurityTab());    break;
    case 'engine':      body.appendChild(buildEngineTab());      break;
    case 'sessions':    body.appendChild(buildSessionsTab());    break;
    case 'wallpapers':  body.appendChild(buildWallpapersTab());  break;
    case 'bookmarklet': body.appendChild(buildBookmarkletTab()); break;
  }
}

function statCard(value, label) {
  const card = el('div', 'stat-card');
  card.appendChild(el('div', 'stat-value', String(value)));
  card.appendChild(el('div', 'stat-label', label));
  return card;
}

function buildProfileTab() {
  const frag = document.createDocumentFragment();
  const u = S.user;

  // Personal stats
  const statsSec = el('div', 'settings-section');
  statsSec.appendChild(el('div', 'settings-section-title', t('profile.stats')));
  const grid = el('div', 'stat-grid mt-1');
  grid.textContent = t('loading');
  GET('/me/stats').then(s => {
    grid.innerHTML = '';
    grid.append(
      statCard(s.bookmarks, t('stat.bookmarks')),
      statCard(s.wallpapers, t('stat.wallpapers')),
      statCard(s.sessions, t('stat.sessions')),
      statCard(new Date(s.member_since * 1000).toLocaleDateString(_locale, { month: 'short', year: 'numeric' }), t('stat.member_since')),
    );
  }).catch(() => { grid.textContent = '—'; });
  statsSec.appendChild(grid);
  frag.appendChild(statsSec);

  const sec = el('div', 'settings-section');
  sec.appendChild(el('div', 'settings-section-title', t('profile.title')));

  const uRow = el('div', 'form-group mt-1');
  const uLbl = el('label', 'form-label', t('profile.username'));
  const uIn  = el('input', 'form-input');
  uIn.type   = 'text'; uIn.value = u.username; uIn.id = 'prof-username';
  uRow.append(uLbl, uIn);

  const eRow = el('div', 'form-group mt-1');
  const eLbl = el('label', 'form-label', t('profile.email'));
  const eIn  = el('input', 'form-input');
  eIn.type   = 'email'; eIn.value = u.email; eIn.id = 'prof-email';
  eRow.append(eLbl, eIn);

  const err  = el('div', 'error-msg'); err.id = 'prof-error';
  const btn  = el('button', 'btn btn-primary mt-1', t('profile.save'));
  btn.onclick = async () => {
    setError('prof-error', '');
    try {
      await PUT('/me', {
        username: $('prof-username').value.trim(),
        email:    $('prof-email').value.trim(),
      });
      S.user = await GET('/me');
    } catch (e) { setError('prof-error', e.message); }
  };

  sec.append(uRow, eRow, err, btn);
  frag.appendChild(sec);

  // Language picker — separate section below the profile form
  const langSec = el('div', 'settings-section');
  langSec.appendChild(el('div', 'settings-section-title', t('profile.locale')));
  const langRow = el('div', 'form-group mt-1');
  const langSel = el('select', 'form-input');
  for (const [code, labelKey] of [['fr', 'profile.locale.fr'], ['en', 'profile.locale.en']]) {
    const opt = document.createElement('option');
    opt.value = code;
    opt.textContent = t(labelKey);
    if (u.locale === code) opt.selected = true;
    langSel.appendChild(opt);
  }
  langSel.onchange = async () => {
    try {
      await PUT('/me/locale', { locale: langSel.value });
      S.user = await GET('/me');
      applyLocale(S.user.locale);
      renderSettingsTab('profile');
    } catch {}
  };
  langRow.appendChild(langSel);
  langSec.appendChild(langRow);
  frag.appendChild(langSec);

  return frag;
}

function buildSecurityTab() {
  const frag = document.createDocumentFragment();

  // Change password
  const pwSec = el('div', 'settings-section');
  pwSec.appendChild(el('div', 'settings-section-title', t('security.password')));

  const curRow = el('div', 'form-group mt-1');
  const curLbl = el('label', 'form-label', t('security.current'));
  const curIn  = el('input', 'form-input');
  curIn.type = 'password'; curIn.id = 'pw-current';
  curRow.append(curLbl, curIn);

  const newRow = el('div', 'form-group mt-1');
  const newLbl = el('label', 'form-label', t('security.new'));
  const newIn  = el('input', 'form-input');
  newIn.type = 'password'; newIn.id = 'pw-new';
  newRow.append(newLbl, newIn);

  const pwErr = el('div', 'error-msg'); pwErr.id = 'pw-error';
  const pwBtn = el('button', 'btn btn-primary mt-1', t('security.change'));
  pwBtn.onclick = async () => {
    setError('pw-error', '');
    try {
      await PUT('/me/password', {
        current_password: $('pw-current').value,
        new_password:     $('pw-new').value,
      });
      $('pw-current').value = '';
      $('pw-new').value     = '';
      setError('pw-error', '✓ ' + t('security.changed'));
    } catch (e) { setError('pw-error', e.message); }
  };

  pwSec.append(curRow, newRow, pwErr, pwBtn);

  // TOTP
  const totpSec = el('div', 'settings-section');
  totpSec.appendChild(el('div', 'settings-section-title', t('security.totp')));
  totpSec.appendChild(buildTOTPSection());

  // Delete account
  const delSec = el('div', 'settings-section delete-account-section');
  delSec.appendChild(el('div', 'settings-section-title', t('account.delete')));
  delSec.appendChild(el('p', 'text-sm text-dim mt-1', t('account.delete.desc')));

  const delPwRow = el('div', 'form-group mt-1');
  const delPwLbl = el('label', 'form-label', t('account.delete.confirm'));
  const delPwIn  = el('input', 'form-input'); delPwIn.type = 'password'; delPwIn.id = 'del-account-pw';
  delPwRow.append(delPwLbl, delPwIn);

  const delErr = el('div', 'error-msg');
  const delBtn = el('button', 'btn btn-danger mt-1', t('account.delete.btn'));
  delBtn.onclick = async () => {
    if (!confirm(t('account.delete.desc'))) return;
    delBtn.disabled = true; delErr.textContent = '';
    try {
      await DEL('/me', { password: delPwIn.value });
      location.reload();
    } catch (e) {
      delErr.textContent = e.message;
      delBtn.disabled = false;
    }
  };
  delSec.append(delPwRow, delErr, delBtn);

  frag.append(pwSec, totpSec, delSec);
  return frag;
}

function buildTOTPSection() {
  const wrap = el('div');

  const note = el('p', 'text-sm text-dim mb-1', t('security.totp.note'));
  wrap.appendChild(note);

  // "Configure" flow: POST /me/totp → show QR → PUT /me/totp with code
  const configBtn = el('button', 'btn btn-secondary', t('security.totp.configure'));
  configBtn.onclick = async () => {
    configBtn.disabled = true;
    try {
      const res = await POST('/me/totp');
      wrap.innerHTML = '';

      const qrWrap = el('div', 'totp-qr-wrap');
      const info   = el('p', 'text-sm text-dim', 'Scannez ce lien avec votre app authenticator :');
      const qrLink = el('div', 'totp-qr-link', res.qr_url);
      const code   = el('input', 'form-input');
      code.placeholder = 'Code à 6 chiffres pour confirmer';
      code.inputMode   = 'numeric';
      code.maxLength   = 6;
      const errEl   = el('div', 'error-msg');
      const confBtn = el('button', 'btn btn-primary mt-1', 'Confirmer');
      confBtn.onclick = async () => {
        try {
          await PUT('/me/totp', { code: code.value.trim() });
          wrap.innerHTML = '';
          wrap.appendChild(el('p', 'text-sm text-dim', 'TOTP activé ✓'));
          wrap.appendChild(buildDisableTOTPBtn());
        } catch (e) { errEl.textContent = e.message; }
      };
      qrWrap.append(info, qrLink, code, errEl, confBtn);
      wrap.append(qrWrap, buildDisableTOTPBtn());
    } catch (e) {
      configBtn.disabled = false;
      alert(e.message);
    }
  };

  wrap.append(configBtn, buildDisableTOTPBtn());
  return wrap;
}

function buildDisableTOTPBtn() {
  const disBtn = el('button', 'btn btn-danger mt-1', t('security.totp.disable'));
  disBtn.onclick = async () => {
    const pw = prompt(t('security.totp.confirm.prompt'));
    if (!pw) return;
    try {
      await DEL('/me/totp', { password: pw });
      renderSettingsTab('security');
    } catch (e) { alert(e.message); }
  };
  return disBtn;
}

function buildEngineTab() {
  const frag = document.createDocumentFragment();
  const sec  = el('div', 'settings-section');
  sec.appendChild(el('div', 'settings-section-title', t('section.engine')));

  const engines = [
    { id: 'duckduckgo', label: 'DuckDuckGo' },
    { id: 'google',     label: 'Google' },
    { id: 'brave',      label: 'Brave' },
    { id: 'bing',       label: 'Bing' },
    { id: 'kagi',       label: 'Kagi' },
    { id: 'custom',     label: t('engine.custom') },
  ];

  const grid = el('div', 'engine-grid');
  for (const eng of engines) {
    const btn = el('button', 'engine-btn', eng.label);
    btn.dataset.engine = eng.id;
    if (S.user.search_engine === eng.id) btn.classList.add('active');
    btn.onclick = async () => {
      let customURL = undefined;
      if (eng.id === 'custom') {
        customURL = prompt(t('engine.custom.prompt'), S.user.search_engine_url || '');
        if (!customURL) return;
      }
      try {
        await PUT('/me/search-engine', { engine: eng.id, custom_url: customURL || undefined });
        S.user = await GET('/me');
        grid.querySelectorAll('.engine-btn').forEach(b => {
          b.classList.toggle('active', b.dataset.engine === eng.id);
        });
      } catch (e) { alert(e.message); }
    };
    grid.appendChild(btn);
  }

  sec.appendChild(grid);
  frag.appendChild(sec);
  return frag;
}

function buildSessionsTab() {
  const frag = document.createDocumentFragment();
  const sec  = el('div', 'settings-section');
  sec.appendChild(el('div', 'settings-section-title', t('sessions.title')));

  const list = el('div'); list.id = 'sessions-list';
  list.textContent = t('loading');

  GET('/me/sessions').then(data => {
    S.sessions = Array.isArray(data) ? data : (data.sessions || []);
    list.innerHTML = '';
    if (!S.sessions.length) {
      list.textContent = t('sessions.none');
      return;
    }
    for (const sess of S.sessions) {
      const row  = el('div', 'session-item');
      const info = el('div', 'session-info');
      const agent = el('div', 'session-agent', sess.user_agent || '—');
      const meta  = el('div', 'session-meta',
        `${sess.ip || '—'} · ${new Date(sess.expires_at * 1000).toLocaleDateString(_locale)}`);

      if (sess.current)        agent.innerHTML += `<span class="badge badge-current">${t('current')}</span>`;
      if (sess.is_bookmarklet) agent.innerHTML += `<span class="badge badge-bookmarklet">${t('sessions.bookmarklet')}</span>`;

      info.append(agent, meta);

      const revokeBtn = el('button', 'btn btn-small btn-danger', t('sessions.revoke'));
      if (sess.current) {
        revokeBtn.disabled = true;
        revokeBtn.title    = t('sessions.current');
      } else {
        revokeBtn.onclick = async () => {
          try {
            await DEL(`/me/sessions/${sess.id}`);
            renderSettingsTab('sessions');
          } catch (e) { alert(e.message); }
        };
      }

      row.append(info, revokeBtn);
      list.appendChild(row);
    }

    const revokeAll = el('button', 'btn btn-danger mt-2', t('sessions.revoke_all'));
    revokeAll.onclick = async () => {
      if (!confirm(t('sessions.revoke.all.confirm'))) return;
      try { await DEL('/me/sessions'); } catch {}
      await logout();
    };
    list.appendChild(revokeAll);
  }).catch(e => { list.textContent = e.message; });

  sec.appendChild(list);
  frag.appendChild(sec);
  return frag;
}

function buildWallpapersTab() {
  const frag = document.createDocumentFragment();
  const sec  = el('div', 'settings-section');
  sec.appendChild(el('div', 'settings-section-title', t('section.wallpapers')));

  // Upload area
  const uploadLabel = el('label', 'upload-area');
  uploadLabel.textContent = t('wp.upload.label');
  const fileIn = el('input', 'sr-only');
  fileIn.type   = 'file';
  fileIn.accept = '.jpg,.jpeg,.png,.webp,.avif,.mp4,.webm';
  fileIn.id     = 'wp-file-input';
  fileIn.setAttribute('for', 'wp-file-input');
  fileIn.onchange = async () => {
    const f = fileIn.files[0];
    if (!f) return;
    const form = new FormData();
    form.append('file', f);
    try {
      const r = await fetch('/api/wallpapers', {
        method:      'POST',
        credentials: 'same-origin',
        body:        form,
      });
      if (!r.ok) {
        const json = await r.json().catch(() => ({}));
        throw new Error(json.error || `Error ${r.status}`);
      }
      S.wallpapers = (await GET('/wallpapers')) || [];
      renderWallpaperGrid(grid);
    } catch (e) { alert(e.message); }
  };
  uploadLabel.appendChild(fileIn);
  sec.appendChild(uploadLabel);

  const grid = el('div', 'wallpaper-grid'); grid.id = 'wp-grid';
  renderWallpaperGrid(grid);
  sec.appendChild(grid);

  frag.appendChild(sec);
  return frag;
}

function renderWallpaperGrid(grid) {
  grid.innerHTML = '';
  if (!S.wallpapers.length) {
    grid.appendChild(el('p', 'text-sm text-dimmer', t('wp.none')));
    return;
  }
  for (const wp of S.wallpapers) {
    const url  = `/media/${S.user.id}/${wp.filename}`;
    const ext  = wp.filename.split('.').pop().toLowerCase();
    const isVid = ['mp4','webm'].includes(ext);

    const thumb    = el('div', 'wp-thumb' + (wp.is_pinned ? ' pinned' : ''));
    const media    = el(isVid ? 'video' : 'img');
    media.src = url;
    if (isVid) { media.muted = true; media.autoplay = true; media.loop = true; }

    const overlay  = el('div', 'wp-thumb-overlay');

    const pinBtn = el('button', 'btn btn-small btn-secondary', wp.is_pinned ? '★' : '☆');
    pinBtn.title = wp.is_pinned ? t('wp.unpin') : t('wp.pin');
    pinBtn.onclick = async () => {
      try {
        await PUT(`/wallpapers/${wp.id}/pin`, { pinned: !wp.is_pinned });
        S.wallpapers = await GET('/wallpapers');
        const g = $('wp-grid');
        if (g) renderWallpaperGrid(g);
      } catch (e) { alert(e.message); }
    };

    const delBtn = el('button', 'btn btn-small btn-danger', '✕');
    delBtn.title  = 'Supprimer';
    delBtn.onclick = async () => {
      if (!confirm(t('wp.delete.confirm'))) return;
      try {
        await DEL(`/wallpapers/${wp.id}`);
        S.wallpapers = await GET('/wallpapers');
        const g = $('wp-grid');
        if (g) renderWallpaperGrid(g);
      } catch (e) { alert(e.message); }
    };

    overlay.append(pinBtn, delBtn);
    thumb.append(media, overlay);
    grid.appendChild(thumb);
  }
}

function buildBookmarkletTab() {
  const frag = document.createDocumentFragment();
  const sec  = el('div', 'settings-section');
  sec.appendChild(el('div', 'settings-section-title', t('section.bookmarklet')));

  const info = el('p', 'text-sm text-dim mb-1', t('bml.desc'));
  sec.appendChild(info);

  const code = el('div', 'bookmarklet-code'); code.id = 'bml-code';
  code.textContent = t('loading');

  const copyBtn  = el('button', 'btn btn-secondary mt-1', t('bml.copy'));
  const revokeBtn = el('button', 'btn btn-danger mt-1',   t('bml.revoke'));

  copyBtn.onclick = () => {
    navigator.clipboard.writeText(code.textContent).then(() => {
      copyBtn.textContent = t('bml.copied');
      setTimeout(() => { copyBtn.textContent = t('bml.copy'); }, 2000);
    });
  };

  revokeBtn.onclick = async () => {
    if (!confirm(t('bml.revoke.confirm'))) return;
    try {
      await DEL('/me/bookmarklet');
      code.textContent = t('bml.revoked');
    } catch (e) { alert(e.message); }
  };

  code.textContent = t('bml.hint');

  const genBtn = el('button', 'btn btn-secondary mt-1', t('bml.generate'));
  genBtn.onclick = async () => {
    try {
      const res = await GET('/me/bookmarklet');
      code.textContent = res.bookmarklet || '—';
    } catch (e) { alert(e.message); }
  };

  sec.append(code, copyBtn, revokeBtn, genBtn);
  frag.appendChild(sec);
  return frag;
}

/* ─── Admin panel ────────────────────────────────────────────────────────── */
function openAdmin() {
  renderAdminTab('stats');
  show('overlay-admin');
}

function closeAdmin() {
  backToHub('overlay-admin');
}

function renderAdminTab(tabName) {
  document.querySelectorAll('#admin-tabs .tab').forEach(t => {
    t.classList.toggle('active', t.dataset.tab === tabName);
  });

  const body = $('admin-body');
  body.innerHTML = '';

  switch (tabName) {
    case 'stats':       body.appendChild(buildAdminStats());       break;
    case 'users':       body.appendChild(buildAdminUsers());       break;
    case 'invitations': body.appendChild(buildAdminInvitations()); break;
    case 'settings':    body.appendChild(buildAdminSettings());    break;
    case 'audit':       body.appendChild(buildAdminAudit());       break;
  }
}

function buildAdminStats() {
  const frag = document.createDocumentFragment();
  const grid = el('div', 'stat-grid');
  grid.textContent = t('loading');

  GET('/admin/stats').then(s => {
    grid.innerHTML = '';
    const stats = [
      { v: s.total_users,      l: t('admin.stat.users') },
      { v: s.active_users,     l: t('admin.stat.active') },
      { v: s.total_bookmarks,  l: t('admin.stat.bookmarks') },
      { v: s.total_wallpapers, l: t('admin.stat.wallpapers') },
      { v: fmtBytes(s.db_size_bytes), l: t('admin.stat.db') },
    ];
    for (const { v, l } of stats) {
      const card = el('div', 'stat-card');
      card.appendChild(el('div', 'stat-value', String(v)));
      card.appendChild(el('div', 'stat-label', l));
      grid.appendChild(card);
    }
  }).catch(e => { grid.textContent = t('error') + ': ' + e.message; });

  frag.appendChild(grid);
  return frag;
}

function fmtBytes(b) {
  if (!b) return '0 B';
  const u = ['B','KB','MB','GB'];
  let i = 0; let n = b;
  while (n >= 1024 && i < u.length - 1) { n /= 1024; i++; }
  return `${n.toFixed(i ? 1 : 0)} ${u[i]}`;
}

function buildAdminUsers() {
  const frag = document.createDocumentFragment();

  const list = el('div'); list.id = 'admin-users-list';
  list.textContent = t('loading');

  GET('/admin/users').then(data => {
    const users = data.users || data.items || data || [];
    list.innerHTML = '';
    for (const u of users) {
      const row = el('div', 'admin-user-row');

      const nameEl = el('div', 'flex-1');
      const name   = el('span', 'user-name', u.username);
      const email  = el('span', 'user-email', ' · ' + u.email);
      if (u.role === 'admin') name.innerHTML += '<span class="badge badge-admin">admin</span>';
      if (!u.is_active)       name.innerHTML += '<span class="badge badge-inactive">suspended</span>';

      // Stats line (bookmarks / wallpapers / sessions)
      const statLine = el('div', 'user-stats', '…');
      // Storage line — updated once stats arrive
      const storageLine = el('div', 'user-stats text-dimmer');

      nameEl.append(name, email, statLine, storageLine);

      GET(`/admin/users/${u.id}/stats`).then(s => {
        statLine.textContent = `${s.bookmarks} ${t('stat.bookmarks').toLowerCase()} · ${s.wallpapers} ${t('stat.wallpapers').toLowerCase()} · ${s.sessions} ${t('stat.sessions').toLowerCase()}`;
        const used      = fmtBytes(s.storage_bytes);
        const quotaVal  = u.storage_quota != null ? fmtBytes(u.storage_quota) : t('admin.upload.global');
        storageLine.textContent = `💾 ${used} / ${quotaVal}`;
        if (u.storage_quota != null && s.storage_bytes > u.storage_quota) {
          storageLine.style.color = 'var(--danger)';
        }
      }).catch(() => { statLine.textContent = ''; });

      const acts = el('div', 'flex gap-1 flex-wrap');

      // Inline upload size limit (single file, MB)
      const limitWrap = el('div', 'flex gap-1 items-center');
      limitWrap.appendChild(el('span', 'text-sm text-dimmer', t('admin.limit.file')));
      const limitInput = el('input', 'form-input form-input-sm');
      limitInput.type = 'number'; limitInput.min = '1'; limitInput.style.width = '5.5rem';
      limitInput.placeholder = t('admin.upload.global');
      limitInput.title = t('admin.upload.limit') + ' (MB)';
      if (u.upload_size_limit != null) limitInput.value = Math.round(u.upload_size_limit / (1024 * 1024));
      const limitSaveBtn = el('button', 'btn btn-small btn-secondary', '💾');
      limitSaveBtn.title = t('admin.upload.limit.save');
      const limitResetBtn = el('button', 'btn btn-small btn-secondary', '↺');
      limitResetBtn.title = t('admin.upload.limit.reset');
      limitSaveBtn.onclick = async () => {
        const mb = limitInput.value.trim();
        if (!mb || isNaN(parseInt(mb, 10)) || parseInt(mb, 10) <= 0) return;
        const limit = parseInt(mb, 10) * 1024 * 1024;
        try { await PUT(`/admin/users/${u.id}/upload-size-limit`, { limit }); renderAdminTab('users'); }
        catch (e) { alert(e.message); }
      };
      limitResetBtn.onclick = async () => {
        try { await PUT(`/admin/users/${u.id}/upload-size-limit`, { limit: null }); renderAdminTab('users'); }
        catch (e) { alert(e.message); }
      };
      limitWrap.append(limitInput, limitSaveBtn, limitResetBtn);

      // Inline storage quota (total, MB)
      const quotaWrap = el('div', 'flex gap-1 items-center');
      quotaWrap.appendChild(el('span', 'text-sm text-dimmer', t('admin.limit.storage')));
      const quotaInput = el('input', 'form-input form-input-sm');
      quotaInput.type = 'number'; quotaInput.min = '1'; quotaInput.style.width = '5.5rem';
      quotaInput.placeholder = t('admin.upload.global');
      quotaInput.title = t('admin.quota') + ' (MB)';
      if (u.storage_quota != null) quotaInput.value = Math.round(u.storage_quota / (1024 * 1024));
      const quotaSaveBtn = el('button', 'btn btn-small btn-secondary', '💾');
      quotaSaveBtn.title = t('admin.quota.save');
      const quotaResetBtn = el('button', 'btn btn-small btn-secondary', '↺');
      quotaResetBtn.title = t('admin.quota.reset');
      quotaSaveBtn.onclick = async () => {
        const mb = quotaInput.value.trim();
        if (!mb || isNaN(parseInt(mb, 10)) || parseInt(mb, 10) <= 0) return;
        const quota = parseInt(mb, 10) * 1024 * 1024;
        try { await PUT(`/admin/users/${u.id}/storage-quota`, { quota }); renderAdminTab('users'); }
        catch (e) { alert(e.message); }
      };
      quotaResetBtn.onclick = async () => {
        try { await PUT(`/admin/users/${u.id}/storage-quota`, { quota: null }); renderAdminTab('users'); }
        catch (e) { alert(e.message); }
      };
      quotaWrap.append(quotaInput, quotaSaveBtn, quotaResetBtn);

      if (u.id !== S.user.id) {
        const suspBtn = el('button', 'btn btn-small btn-secondary',
          u.is_active ? t('admin.user.suspend') : t('admin.user.activate'));
        suspBtn.onclick = async () => {
          try {
            if (u.is_active) await PUT(`/admin/users/${u.id}/suspend`);
            else             await PUT(`/admin/users/${u.id}/activate`);
            renderAdminTab('users');
          } catch (e) { alert(e.message); }
        };

        const delBtn = el('button', 'btn btn-small btn-danger', t('admin.user.delete'));
        delBtn.title = t('admin.delete.warning');
        delBtn.onclick = async () => {
          if (!confirm(`${t('admin.delete.warning')}\n\n${t('admin.user.delete.confirm').replace('{u}', u.username)}`)) return;
          try {
            await DEL(`/admin/users/${u.id}`);
            renderAdminTab('users');
          } catch (e) { alert(e.message); }
        };

        acts.append(suspBtn, delBtn);
      } else {
        acts.appendChild(el('span', 'text-sm text-dimmer', t('admin.user.you')));
      }

      acts.append(limitWrap, quotaWrap);

      row.append(nameEl, acts);
      list.appendChild(row);
    }
  }).catch(e => { list.textContent = t('error') + ': ' + e.message; });

  frag.append(list);
  return frag;
}

const AUDIT_LABELS = {
  login:                            'Connexion',
  login_sso:                        'Connexion SSO',
  login_failed:                     'Échec de connexion',
  logout:                           'Déconnexion',
  password_change:                  'Mot de passe modifié',
  totp_enabled:                     'TOTP activé',
  totp_disabled:                    'TOTP désactivé',
  user_created:                     'Compte créé',
  user_created_sso:                 'Compte créé (SSO)',
  user_deleted:                     'Compte supprimé',
  user_suspended:                   'Compte suspendu',
  register_blocked_duplicate_email: 'Inscription bloquée (email déjà utilisé)',
  registration_requested:           'Demande d’inscription',
  registration_completed:           'Inscription finalisée',
  registration_revoked:             'Demande d’inscription révoquée',
  invitation_sent:                  'Invitation envoyée',
  invitation_revoked:               'Invitation révoquée',
  bookmark_import:                  'Import de marque-pages',
  wallpaper_upload:                 'Fond d’écran ajouté',
  wallpaper_delete:                 'Fond d’écran supprimé',
};
function auditLabel(action) {
  return AUDIT_LABELS[action] || action.replace(/_/g, ' ');
}

function buildAdminAudit() {
  const frag  = document.createDocumentFragment();
  const list  = el('div'); list.id = 'admin-audit-list';
  list.textContent = t('loading');

  GET('/admin/audit?limit=100').then(data => {
    const entries = data.entries || data.items || data || [];
    list.innerHTML = '';
    if (!entries.length) { list.textContent = t('audit.none'); return; }
    for (const e of entries) {
      const row    = el('div', 'audit-row');
      const action = el('span', 'audit-action', auditLabel(e.action));
      const ip     = el('span', 'audit-ip', e.ip || '—');
      const user   = el('span', 'audit-user', e.username || (e.user_id ? '—' : 'system'));
      const time   = el('span', 'audit-time',
        new Date(e.created_at * 1000).toLocaleString(_locale, { dateStyle: 'short', timeStyle: 'short' }));
      row.append(action, ip, user, time);
      list.appendChild(row);
    }
  }).catch(e => { list.textContent = t('error') + ': ' + e.message; });

  frag.appendChild(list);
  return frag;
}

function buildAdminSettings() {
  const frag = document.createDocumentFragment();

  const title = el('div', 'settings-section-title', t('admin.menu.label'));
  const desc = el('p', 'text-sm text-dim mb-1', t('admin.menu.desc'));

  const row = el('div', 'flex gap-1 mb-1');
  const input = el('input', 'form-input flex-1');
  input.type = 'text'; input.placeholder = '!menu';
  const saveBtn = el('button', 'btn btn-primary', t('btn.save'));
  const msg = el('div', 'error-msg');
  row.append(input, saveBtn);

  GET('/admin/settings/menu').then(s => {
    input.value = s.menu_bang;
    if (s.locked) {
      input.disabled = true;
      saveBtn.disabled = true;
      msg.className = 'text-sm text-dimmer';
      msg.textContent = t('admin.menu.locked');
    }
  }).catch(e => { msg.textContent = e.message; });

  saveBtn.onclick = async () => {
    msg.className = 'error-msg'; msg.textContent = '';
    let v = input.value.trim();
    if (v && v[0] !== '!') v = '!' + v;
    try {
      const r = await PUT('/admin/settings/menu', { menu_bang: v });
      S.menuBang = r.menu_bang;
      input.value = r.menu_bang;
      msg.className = 'text-sm text-dim';
      msg.textContent = t('admin.saved');
    } catch (e) { msg.className = 'error-msg'; msg.textContent = e.message; }
  };

  frag.append(title, desc, row, msg);
  frag.append(buildAdminSSO());
  frag.append(buildAdminSystem());
  return frag;
}

function buildAdminSystem() {
  const wrap = el('div');
  wrap.style.marginTop = '2rem';
  wrap.appendChild(el('div', 'settings-section-title', t('admin.section.system')));
  wrap.appendChild(el('p', 'text-sm text-dim mb-1',
    'Settings locked in compose are greyed out. Secret values are never shown.'));

  // Editable runtime settings
  const mkNum = (labelTxt, ph) => {
    const g = el('div', 'form-group'); g.appendChild(el('label', 'form-label', labelTxt));
    const i = el('input', 'form-input'); i.type = 'number'; i.placeholder = ph || ''; g.appendChild(i);
    return { g, i };
  };
  const fTotp = (() => {
    const g = el('div', 'form-group'); g.appendChild(el('label', 'form-label', 'TOTP issuer name'));
    const i = el('input', 'form-input'); i.type = 'text'; g.appendChild(i);
    return { g, i };
  })();
  const fWp = mkNum('Default wallpaper limit', '10');
  const fBm = mkNum('Bookmarklet token duration (days)', '90');
  const saveBtn = el('button', 'btn btn-primary', t('btn.save'));
  const msg = el('div', 'error-msg');

  // SMTP (editable unless env-locked)
  const smtpWrap = el('div'); smtpWrap.style.marginTop = '1.6rem';
  const fHost = (() => { const g = el('div','form-group'); g.appendChild(el('label','form-label','SMTP server')); const i = el('input','form-input'); i.placeholder='smtp.example.com'; g.appendChild(i); return {g,i}; })();
  const fPort = mkNum('Port', '587');
  const fUser = (() => { const g = el('div','form-group'); g.appendChild(el('label','form-label','Username')); const i = el('input','form-input'); g.appendChild(i); return {g,i}; })();
  const fPass = (() => { const g = el('div','form-group'); g.appendChild(el('label','form-label','Password')); const i = el('input','form-input'); i.type='password'; i.placeholder='Leave blank to keep'; g.appendChild(i); return {g,i}; })();
  const fFrom = (() => { const g = el('div','form-group'); g.appendChild(el('label','form-label','From address')); const i = el('input','form-input'); i.type='email'; i.placeholder='cairn@example.com'; g.appendChild(i); return {g,i}; })();
  const fTls  = (() => {
    const g = el('div', 'effect-row');
    g.appendChild(el('div', 'effect-label', 'TLS (STARTTLS)'));
    const sw = el('label', 'toggle-switch');
    const i = el('input'); i.type = 'checkbox';
    sw.append(i, el('span', 'toggle-track'));
    g.appendChild(sw);
    return {g,i};
  })();
  const smtpStatus = el('div', 'text-sm text-dimmer mb-1');
  const smtpMsg = el('div', 'error-msg');

  // Read-only infrastructure
  const roWrap = el('div'); roWrap.style.marginTop = '1.6rem';

  const lockField = (f, locked) => { if (locked) { f.i.disabled = true; f.i.title = 'Défini dans le compose'; } };

  GET('/admin/settings/system').then(s => {
    fTotp.i.value = s.editable.totp_issuer.value;     lockField(fTotp, s.editable.totp_issuer.locked);
    fWp.i.value   = s.editable.wallpaper_limit.value;  lockField(fWp,   s.editable.wallpaper_limit.locked);
    fBm.i.value   = s.editable.bookmarklet_days.value; lockField(fBm,   s.editable.bookmarklet_days.locked);

    // SMTP
    fHost.i.value = s.smtp.host || '';
    fPort.i.value = s.smtp.port || 587;
    fUser.i.value = s.smtp.user || '';
    fFrom.i.value = s.smtp.from || '';
    fTls.i.checked = !!s.smtp.tls;
    if (s.smtp.has_password) fPass.i.placeholder = '•••••••• (leave blank to keep)';
    smtpStatus.textContent = s.smtp.configured ? '● Email configured' : '○ Email not configured — invitations and password resets will not work';
    smtpStatus.style.color = s.smtp.configured ? 'var(--success)' : 'var(--danger)';
    if (s.smtp.locked) {
      [fHost,fPort,fUser,fPass,fFrom,fTls].forEach(f => f.i.disabled = true);
      smtpSaveBtn.disabled = true;
      smtpMsg.className = 'text-sm text-dimmer';
      smtpMsg.textContent = 'Configured via compose (CAIRN_SMTP_*) — not editable here.';
    }

    roWrap.innerHTML = '';
    roWrap.appendChild(el('div', 'settings-section-title', t('admin.section.infra')));
    const rows = [
      [t('admin.sys.addr'),           s.system.addr],
      [t('admin.sys.env'),            s.system.env],
      [t('admin.sys.base_url'),       s.system.base_url],
      [t('admin.sys.db_path'),        s.system.db_path],
      [t('admin.sys.media_path'),     s.system.media_path],
      [t('admin.sys.max_upload'),     fmtBytes(s.system.max_upload_size)],
      [t('admin.sys.storage_quota'),  fmtBytes(s.system.default_storage_quota)],
      [t('admin.sys.trusted_proxy'),  s.system.trusted_proxy ? t('admin.sys.yes') : t('admin.sys.no')],
      [t('admin.sys.session_secret'), s.system.session_secret ? `•••••••• ${t('admin.sys.set')}` : t('admin.sys.not_set')],
    ];
    for (const [k, v] of rows) {
      const r = el('div', 'sysinfo-row');
      r.append(el('span', 'sysinfo-key', k), el('span', 'sysinfo-val', String(v)));
      roWrap.appendChild(r);
    }
  }).catch(e => { msg.textContent = e.message; });

  saveBtn.onclick = async () => {
    msg.className = 'error-msg'; msg.textContent = '';
    try {
      await PUT('/admin/settings/system', {
        totp_issuer:      fTotp.i.value.trim(),
        wallpaper_limit:  parseInt(fWp.i.value, 10) || 0,
        bookmarklet_days: parseInt(fBm.i.value, 10) || 0,
      });
      msg.className = 'text-sm text-dim'; msg.textContent = t('admin.saved');
    } catch (e) { msg.className = 'error-msg'; msg.textContent = e.message; }
  };

  const smtpSaveBtn = el('button', 'btn btn-primary', t('btn.save') + ' SMTP');
  smtpSaveBtn.onclick = async () => {
    smtpMsg.className = 'error-msg'; smtpMsg.textContent = '';
    try {
      const r = await PUT('/admin/settings/system', { smtp: {
        host: fHost.i.value.trim(),
        port: parseInt(fPort.i.value, 10) || 587,
        user: fUser.i.value.trim(),
        pass: fPass.i.value,
        from: fFrom.i.value.trim(),
        tls:  fTls.i.checked,
      }});
      fPass.i.value = '';
      smtpStatus.textContent = r.smtp.configured ? '● Email configured' : '○ Email not configured';
      smtpStatus.style.color = r.smtp.configured ? 'var(--success)' : 'var(--danger)';
      // Refresh the cached user flag used by the setup banner.
      try { S.user = await GET('/me'); } catch {}
      smtpMsg.className = 'text-sm text-dim'; smtpMsg.textContent = t('admin.saved');
    } catch (e) { smtpMsg.className = 'error-msg'; smtpMsg.textContent = e.message; }
  };

  smtpWrap.append(el('div','settings-section-title', t('admin.section.smtp')), smtpStatus,
    fHost.g, fPort.g, fUser.g, fPass.g, fFrom.g, fTls.g, smtpSaveBtn, smtpMsg);

  wrap.append(fTotp.g, fWp.g, fBm.g, saveBtn, msg, smtpWrap, roWrap);
  return wrap;
}

function buildAdminSSO() {
  const wrap = el('div');
  wrap.style.marginTop = '2rem';
  const title = el('div', 'settings-section-title', t('admin.section.sso'));
  const desc = el('p', 'text-sm text-dim mb-1',
    t('admin.sso.desc') + ' — Redirect URI: ' + location.origin + '/api/auth/sso/callback');

  const mkField = (labelTxt, ph, type) => {
    const g = el('div', 'form-group');
    g.appendChild(el('label', 'form-label', labelTxt));
    const i = el('input', 'form-input');
    i.type = type || 'text'; i.placeholder = ph || '';
    g.appendChild(i);
    return { g, i };
  };

  const fName   = mkField('Display name', 'Authentik');
  const fIssuer = mkField('Issuer URL', 'https://auth.example.com/application/o/cairn/');
  const fClient = mkField('Client ID', '');
  const fSecret = mkField('Client Secret', 'Leave blank to keep', 'password');
  const fScopes = mkField('Scopes', 'openid profile email');
  const saveBtn = el('button', 'btn btn-primary', 'Save SSO');
  const msg = el('div', 'error-msg');
  const status = el('div', 'text-sm text-dimmer mb-1');

  GET('/admin/settings/sso').then(s => {
    fName.i.value   = s.provider_name || '';
    fIssuer.i.value = s.issuer || '';
    fClient.i.value = s.client_id || '';
    fScopes.i.value = s.scopes || '';
    if (s.has_secret) fSecret.i.placeholder = '•••••••• (leave blank to keep)';
    status.textContent = s.enabled ? '● SSO active' : '○ SSO inactive';
    status.style.color = s.enabled ? 'var(--success)' : 'var(--text-dimmer)';
    if (s.locked) {
      [fName, fIssuer, fClient, fSecret, fScopes].forEach(f => f.i.disabled = true);
      saveBtn.disabled = true;
      msg.className = 'text-sm text-dimmer';
      msg.textContent = 'Configured via compose (CAIRN_OIDC_*) — not editable here.';
    }
  }).catch(e => { msg.textContent = e.message; });

  saveBtn.onclick = async () => {
    msg.className = 'error-msg'; msg.textContent = '';
    try {
      const r = await PUT('/admin/settings/sso', {
        provider_name: fName.i.value.trim(),
        issuer:        fIssuer.i.value.trim(),
        client_id:     fClient.i.value.trim(),
        client_secret: fSecret.i.value,
        scopes:        fScopes.i.value.trim(),
      });
      fSecret.i.value = '';
      if (r.has_secret) fSecret.i.placeholder = '•••••••• (leave blank to keep)';
      status.textContent = r.enabled ? '● SSO active' : '○ SSO inactive';
      status.style.color = r.enabled ? 'var(--success)' : 'var(--text-dimmer)';
      msg.className = 'text-sm text-dim';
      msg.textContent = t('admin.saved');
    } catch (e) { msg.className = 'error-msg'; msg.textContent = e.message; }
  };

  wrap.append(title, desc, status, fName.g, fIssuer.g, fClient.g, fSecret.g, fScopes.g, saveBtn, msg);
  return wrap;
}

function buildAdminInvitations() {
  const frag = document.createDocumentFragment();

  // ── Registration toggle ──────────────────────────────────────────────────
  const regSection = el('div', 'settings-section');
  const toggleErr = el('div', 'error-msg');

  const regRow = el('div', 'effect-row');
  const regLabelWrap = el('div', 'effect-label-wrap');
  regLabelWrap.appendChild(el('div', 'effect-label', t('section.registration')));
  regLabelWrap.appendChild(el('div', 'effect-sub', t('admin.reg.desc')));
  regRow.appendChild(regLabelWrap);

  const regSwitch = el('label', 'toggle-switch');
  const regInp = el('input'); regInp.type = 'checkbox'; regInp.disabled = true;
  const regTrack = el('span', 'toggle-track');
  regSwitch.append(regInp, regTrack);
  regRow.appendChild(regSwitch);

  GET('/admin/settings/registration').then(s => {
    regInp.checked  = s.open_registration;
    regInp.disabled = false;
  }).catch(e => {
    toggleErr.textContent = e.message;
  });

  regInp.onchange = async () => {
    regInp.disabled = true;
    toggleErr.textContent = '';
    try {
      const res = await PUT('/admin/settings/registration', { open_registration: regInp.checked });
      regInp.checked  = res.open_registration;
    } catch (e) {
      regInp.checked = !regInp.checked;
      toggleErr.textContent = e.message || 'Error';
    } finally {
      regInp.disabled = false;
    }
  };

  regSection.append(regRow, toggleErr);
  frag.appendChild(regSection);

  // ── Pending open-registration requests ───────────────────────────────────
  const pendingSection = el('div', 'settings-section');
  pendingSection.appendChild(el('div', 'settings-section-title', t('admin.pending.title')));
  const pendingList = el('div');
  pendingList.textContent = t('loading');
  pendingSection.appendChild(pendingList);

  GET('/admin/pending-registrations').then(prs => {
    pendingList.innerHTML = '';
    if (!prs.length) {
      pendingList.appendChild(el('p', 'text-sm text-dimmer', t('admin.pending.none')));
      return;
    }
    for (const pr of prs) {
      const row = el('div', 'admin-user-row');

      const info = el('div');
      const name = el('span', 'user-name', pr.username);
      let badge = '';
      if (pr.completed)    badge = `<span class="badge badge-inactive">${t('admin.pending.completed')}</span>`;
      else if (pr.expired) badge = `<span class="badge badge-inactive">${t('admin.inv.expired')}</span>`;
      else                 badge = `<span class="badge badge-admin">${t('admin.inv.pending')}</span>`;
      name.innerHTML += badge;

      const detail = el('span', 'user-email',
        ' · ' + pr.email + ' · ' + t('admin.inv.expires') + ' ' +
        new Date(pr.expires_at * 1000).toLocaleString(_locale, { dateStyle: 'short', timeStyle: 'short' }));
      info.append(name, detail);

      const acts = el('div', 'flex gap-1');
      const revokeBtn = el('button', 'btn btn-small btn-danger', t('admin.pending.revoke'));
      revokeBtn.onclick = async () => {
        try { await DEL(`/admin/pending-registrations/${pr.id}`); renderAdminTab('invitations'); }
        catch (e) { alert(e.message); }
      };
      acts.appendChild(revokeBtn);

      row.append(info, acts);
      pendingList.appendChild(row);
    }
  }).catch(e => { pendingList.textContent = t('error') + ': ' + e.message; });

  frag.appendChild(pendingSection);

  // ── Invite form ──────────────────────────────────────────────────────────
  const inviteSection = el('div', 'settings-section');
  const form = el('div', 'flex gap-1 mb-1');
  const emailInput = el('input', 'form-input flex-1');
  emailInput.type = 'email'; emailInput.placeholder = 'user@example.com';
  const inviteBtn = el('button', 'btn btn-primary', t('admin.inv.invite'));
  const inviteErr = el('div', 'error-msg');
  form.append(emailInput, inviteBtn);

  inviteBtn.onclick = async () => {
    inviteErr.textContent = '';
    try {
      await POST('/admin/invitations', { email: emailInput.value.trim() });
      emailInput.value = '';
      renderAdminTab('invitations');
    } catch (e) { inviteErr.textContent = e.message; }
  };

  // Invitation list
  const list = el('div');
  list.textContent = t('loading');

  GET('/admin/invitations').then(invs => {
    list.innerHTML = '';
    if (!invs.length) { list.appendChild(el('p', 'text-sm text-dimmer', t('admin.inv.none'))); return; }
    for (const inv of invs) {
      const row = el('div', 'admin-user-row');

      const info = el('div');
      const email = el('span', 'user-name', inv.email);
      let badge = '';
      if (inv.used)         badge = `<span class="badge badge-inactive">${t('admin.inv.used')}</span>`;
      else if (inv.expired) badge = `<span class="badge badge-inactive">${t('admin.inv.expired')}</span>`;
      else                  badge = `<span class="badge badge-admin">${t('admin.inv.pending')}</span>`;
      email.innerHTML += badge;

      const exp = el('span', 'user-email',
        ' · ' + t('admin.inv.expires') + ' ' + new Date(inv.expires_at * 1000).toLocaleString(_locale, { dateStyle: 'short', timeStyle: 'short' }));
      info.append(email, exp);

      const acts = el('div', 'flex gap-1');

      if (inv.pending) {
        const resendBtn = el('button', 'btn btn-small btn-secondary', t('admin.inv.resend'));
        resendBtn.onclick = async () => {
          try { await POST(`/admin/invitations/${inv.id}/resend`); renderAdminTab('invitations'); }
          catch (e) { alert(e.message); }
        };
        acts.appendChild(resendBtn);
      }

      const delBtn = el('button', 'btn btn-small btn-danger', t('admin.inv.revoke'));
      delBtn.onclick = async () => {
        try { await DEL(`/admin/invitations/${inv.id}`); renderAdminTab('invitations'); }
        catch (e) { alert(e.message); }
      };
      acts.appendChild(delBtn);

      row.append(info, acts);
      list.appendChild(row);
    }
  }).catch(e => { list.textContent = t('error') + ': ' + e.message; });

  inviteSection.append(form, inviteErr, list);
  frag.appendChild(inviteSection);
  return frag;
}

/* ─── New user (admin) ───────────────────────────────────────────────────── */
async function createUser() {
  setError('nu-error', '');
  const username = $('nu-username').value.trim();
  const email    = $('nu-email').value.trim();
  const password = $('nu-password').value;

  try {
    await POST('/auth/register', { username, email, password });
    hide('modal-new-user');
    renderAdminTab('users');
  } catch (e) { setError('nu-error', e.message); }
}

/* ─── SSO ────────────────────────────────────────────────────────────────── */
async function checkSSO() {
  // Surface a provider error returned via the callback redirect.
  const params = new URLSearchParams(window.location.search);
  if (params.get('sso_error')) {
    setError('login-error', 'SSO login failed. Try again or use your credentials.');
    window.history.replaceState({}, '', '/');
  }
  try {
    const cfg = await GET('/auth/sso/config');
    if (cfg.enabled) {
      $('sso-provider').textContent = cfg.provider_name || 'SSO';
      show('sso-block');
    }
  } catch { /* SSO not configured — ignore */ }
}

/* ─── Public auth config (open registration) ────────────────────────────── */
async function checkPublicConfig() {
  try {
    const cfg = await GET('/auth/config');
    toggle('register-link', !!cfg.open_registration);
  } catch { /* ignore — keep link hidden */ }
}

/* ─── Setup page (invite token or pending-registration token) ─────────────── */
// Render the TOTP setup block: open-in-authenticator link + copyable secret.
// No external QR library needed — the otpauth:// URI opens the authenticator app directly
// on mobile; desktop users copy the secret and enter it manually.
function drawQR(container, totpURI, secret) {
  container.innerHTML = '';

  // "Open in authenticator" button — works natively on iOS/Android
  const openBtn = el('a', 'btn btn-secondary btn-full');
  openBtn.href = totpURI;
  openBtn.textContent = '📱 ' + t('setup.open.app');
  container.appendChild(openBtn);

  // Copyable secret for desktop manual entry
  const row = el('div', 'setup-secret-row mt-1');
  const secretSpan = el('span', '', secret);
  secretSpan.style.flex = '1';
  const copyBtn = el('button', '', '⎘');
  copyBtn.title = t('btn.copy') || 'Copy';
  copyBtn.onclick = () => {
    navigator.clipboard.writeText(secret).then(() => {
      copyBtn.textContent = '✓';
      setTimeout(() => { copyBtn.textContent = '⎘'; }, 1500);
    });
  };
  row.append(secretSpan, copyBtn);
  container.appendChild(row);
}

// State for the setup page
const _setup = { mode: null, token: null, totpURI: null, totpSecret: null };

async function openSetupPage(mode, token) {
  _setup.mode  = mode; // 'reg' or 'invite'
  _setup.token = token;
  _setup.totpURI = null;
  _setup.totpSecret = null;

  // Reset UI
  hide('setup-step-password');
  show('setup-step-totp');
  hide('setup-totp-wrap');
  setError('setup-step1-error', '');
  setError('setup-step2-error', '');
  $('setup-qr').innerHTML = '';
  $('setup-password').value  = '';
  $('setup-password2').value = '';

  if (mode === 'reg') {
    hide('setup-username-group');
    // Load data immediately — token carries username/email/totp
    try {
      const data = await GET(`/auth/setup?token=${encodeURIComponent(token)}`);
      $('setup-greeting').textContent = `${data.username} · ${data.email}`;
      _setup.totpURI    = data.totp_uri;
      _setup.totpSecret = data.totp_secret;
      $('setup-next-btn').textContent = t('setup.scan');
      show('setup-totp-wrap');
      drawQR($('setup-qr'), data.totp_uri, data.totp_secret);
      $('setup-totp-secret').textContent = '';
    } catch {
      hide('view-setup');
      show('view-login');
      setError('login-error', t('setup.invalid'));
      return;
    }
  } else {
    // Invite flow — need username first
    $('setup-greeting').textContent = '';
    $('setup-username').value = '';
    show('setup-username-group');
    $('setup-next-btn').textContent = t('setup.next');
    $('setup-username').focus();
  }

  hide('view-login');
  show('view-setup');
  window.history.replaceState({}, '', '/');
}

async function setupNext() {
  setError('setup-step1-error', '');

  if (_setup.mode === 'invite' && !_setup.totpURI) {
    // Phase 1: submit username to get TOTP
    const username = $('setup-username').value.trim();
    if (!username) { setError('setup-step1-error', t('register.username') + ' required'); return; }
    const btn = $('setup-next-btn'); btn.disabled = true;
    try {
      const data = await POST(`/auth/invite/${_setup.token}/prepare`, { username });
      _setup.totpURI    = data.totp_uri;
      _setup.totpSecret = data.totp_secret;
      $('setup-greeting').textContent = `${username} · ${data.email}`;
      hide('setup-username-group');
      show('setup-totp-wrap');
      drawQR($('setup-qr'), data.totp_uri, data.totp_secret);
      $('setup-totp-secret').textContent = '';
      $('setup-next-btn').textContent = t('setup.scan');
      $('setup-totp-code').focus();
    } catch (e) {
      setError('setup-step1-error', e.message);
    } finally { btn.disabled = false; }
    return;
  }

  // Phase 2: validate TOTP code entered, move to password step
  const code = $('setup-totp-code').value.trim();
  if (code.length !== 6 || !/^\d{6}$/.test(code)) {
    setError('setup-step1-error', t('login.totp') + ': 6 digits required');
    return;
  }
  hide('setup-step-totp');
  show('setup-step-password');
  $('setup-password').focus();
}

async function setupFinish() {
  setError('setup-step2-error', '');
  const pw  = $('setup-password').value;
  const pw2 = $('setup-password2').value;
  if (pw !== pw2) { setError('setup-step2-error', t('setup.password.mismatch')); return; }
  if (pw.length < 12) { setError('setup-step2-error', t('register.password.hint')); return; }

  const code = $('setup-totp-code').value.trim();
  const btn  = $('setup-finish-btn'); btn.disabled = true;
  try {
    if (_setup.mode === 'reg') {
      await POST('/auth/complete-setup', { token: _setup.token, password: pw, totp_code: code });
    } else {
      await POST(`/auth/invite/${_setup.token}/complete`, { password: pw, totp_code: code });
    }
    hide('view-setup');
    await boot();
  } catch (e) {
    // TOTP code might be stale — go back to step 1
    setError('setup-step2-error', e.message);
    btn.disabled = false;
  }
}

async function checkSetupOrInviteToken() {
  const params = new URLSearchParams(window.location.search);
  const setupToken  = params.get('token');
  const inviteToken = params.get('invite');

  if (setupToken) {
    await openSetupPage('reg', setupToken);
    return true;
  }
  if (inviteToken) {
    await openSetupPage('invite', inviteToken);
    return true;
  }
  return false;
}

/* ─── Boot ───────────────────────────────────────────────────────────────── */
async function boot() {
  try {
    S.user = await GET('/me');
  } catch {
    hide('view-home');
    hide('view-setup');
    show('view-login');
    await checkSSO();
    await checkPublicConfig();
    const onSetup = await checkSetupOrInviteToken();
    if (onSetup) return;
    $('login-email').focus();
    return;
  }

  // Load wallpapers
  try {
    S.wallpapers = (await GET('/wallpapers')) || [];
  } catch { S.wallpapers = []; }

  // Apply user locale before rendering any UI text.
  applyLocale(S.user.locale || 'en');

  // Apply saved theme preferences (blur, rain)
  applyThemePrefs(loadThemePrefs());

  // Configurable menu bang + placeholder hint
  if (S.user.menu_bang) S.menuBang = S.user.menu_bang;
  const si = $('search-input');
  if (si) si.placeholder = `${t('search.placeholder')}  ${S.menuBang}`;

  loadWallpaper();
  startClock();

  hide('view-login');
  show('view-home');
  $('search-input').focus();
}

/* ─── Event wiring ───────────────────────────────────────────────────────── */
document.addEventListener('DOMContentLoaded', () => {
  initRain();
  initDust();

  // Boot
  boot();

  // Login
  $('login-form').addEventListener('submit', handleLogin);
  $('forgot-form').addEventListener('submit', handleForgot);
  $('register-form').addEventListener('submit', handleRegister);
  const _regValidate = () => {
    const username = $('reg-username').value.trim();
    const email    = $('reg-email').value.trim();
    const emailOk  = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
    $('reg-submit').disabled = username.length < 2 || !emailOk;
  };
  $('reg-username').addEventListener('input', _regValidate);
  $('reg-email').addEventListener('input', _regValidate);
  _regValidate(); // initial state: button disabled
  $('forgot-link').addEventListener('click', showForgotForm);
  $('forgot-back').addEventListener('click', showLoginForm);
  $('register-link').addEventListener('click', showRegisterForm);
  $('reg-back').addEventListener('click', showLoginForm);
  $('reg-back-2').addEventListener('click', showLoginForm);

  // Setup page
  $('setup-next-btn').addEventListener('click', setupNext);
  $('setup-finish-btn').addEventListener('click', setupFinish);
  $('setup-back-btn').addEventListener('click', () => {
    hide('setup-step-password');
    show('setup-step-totp');
    setError('setup-step2-error', '');
  });
  $('setup-totp-code').addEventListener('keydown', e => {
    if (e.key === 'Enter') { e.preventDefault(); setupNext(); }
  });
  $('setup-password2').addEventListener('keydown', e => {
    if (e.key === 'Enter') { e.preventDefault(); setupFinish(); }
  });

  // Home + hub
  $('search-form').addEventListener('submit', handleSearch);
  initSearchSuggestions();

  // Spotlight: dim background only when there is text in the search bar.
  const _si = $('search-input');
  const _so = $('spotlight-overlay');
  function updateSearching() {
    document.body.classList.toggle('searching', _si.value.trim().length > 0);
  }
  function exitSearching() { document.body.classList.remove('searching'); }
  _si.addEventListener('input', updateSearching);
  _si.addEventListener('blur',  exitSearching);
  _so.addEventListener('click', () => { exitSearching(); _si.value = ''; _si.blur(); });
  $('hub-close').addEventListener('click', closeHub);
  $('hub-overlay').addEventListener('mousedown', e => { if (e.target === $('hub-overlay')) closeHub(); });
  $('tile-bookmarks').addEventListener('click', () => { closeHub(); openBookmarks(); });
  $('tile-settings').addEventListener('click', () => { closeHub(); openSettings(); });
  $('tile-theme').addEventListener('click', () => { closeHub(); openTheme(); });
  $('tile-admin').addEventListener('click', () => { closeHub(); openAdmin(); });
  $('tile-logout').addEventListener('click', () => { closeHub(); logout(); });
  $('hub-setup').addEventListener('click', () => { closeHub(); openAdmin(); renderAdminTab('settings'); });

  // Theme overlay
  $('theme-close').addEventListener('click', closeTheme);
  $('overlay-theme').addEventListener('mousedown', e => { if (e.target === $('overlay-theme')) closeTheme(); });
  document.querySelectorAll('#theme-tabs .tab').forEach(tab => {
    tab.addEventListener('click', () => renderThemeTab(tab.dataset.tab));
  });

  // Bookmark overlay
  $('bm-close').addEventListener('click', closeBookmarks);
  $('bm-add-btn').addEventListener('click', openAddBookmark);
  $('bm-export-btn').addEventListener('click', exportBookmarks);
  $('bm-import-file').addEventListener('change', e => importBookmarks(e.target.files[0]));
  $('bm-search-input').addEventListener('input', () => {
    S.bmFilter = $('bm-search-input').value.trim();
    S.bmOffset = 0;
    loadBookmarks();
  });
  $('bm-folder-filter').addEventListener('change', () => {
    S.bmFolder = $('bm-folder-filter').value;
    S.bmOffset = 0;
    loadBookmarks();
  });

  // Bookmark modal
  $('modal-bm-save').addEventListener('click', saveBookmark);
  $('modal-bm-cancel').addEventListener('click', () => hide('modal-bookmark'));

  // Settings overlay
  $('settings-close').addEventListener('click', closeSettings);
  document.querySelectorAll('#settings-tabs .tab').forEach(tab => {
    tab.addEventListener('click', () => renderSettingsTab(tab.dataset.tab));
  });

  // Admin overlay
  $('admin-close').addEventListener('click', closeAdmin);
  document.querySelectorAll('#admin-tabs .tab').forEach(tab => {
    tab.addEventListener('click', () => renderAdminTab(tab.dataset.tab));
  });

  // New user modal
  $('nu-save').addEventListener('click', createUser);
  $('nu-cancel').addEventListener('click', () => hide('modal-new-user'));

  // Backdrop click: panels step back to the hub; modals just close.
  ['overlay-bookmarks','overlay-settings','overlay-admin'].forEach(id => {
    $(id).addEventListener('click', e => { if (e.target === $(id)) backToHub(id); });
  });
  ['modal-bookmark','modal-new-user'].forEach(id => {
    $(id).addEventListener('click', e => { if (e.target === $(id)) hide(id); });
  });

  // Keyboard: Escape closes the top-most layer (modals → panels → hub).
  document.addEventListener('keydown', e => {
    if (e.key !== 'Escape') return;
    for (const id of ['modal-bookmark','modal-new-user']) {
      if (!$(id).classList.contains('hidden')) { hide(id); return; }
    }
    for (const id of ['overlay-bookmarks','overlay-settings','overlay-admin']) {
      if (!$(id).classList.contains('hidden')) { backToHub(id); return; }
    }
    if (!$('hub-overlay').classList.contains('hidden')) closeHub();
  });
});
