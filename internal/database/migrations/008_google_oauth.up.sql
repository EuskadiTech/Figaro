-- Add Google OAuth settings to system_settings table
INSERT INTO system_settings (key, value, category, description) VALUES
    ('google_oauth_enabled', 'false', 'oauth', 'Activar login con Google'),
    ('google_client_id', '', 'oauth', 'ID de cliente de Google OAuth'),
    ('google_client_secret', '', 'oauth', 'Secreto de cliente de Google OAuth'),
    ('google_redirect_url', '', 'oauth', 'URL de redirección OAuth (debe coincidir con Google Console)'),
    ('google_hosted_domain', '', 'oauth', 'Dominio específico para restringir login (opcional)');