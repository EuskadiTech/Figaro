-- Create system settings table to store configuration values
CREATE TABLE system_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT NOT NULL UNIQUE,
    value TEXT NOT NULL,
    category TEXT NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create index on category for efficient filtering
CREATE INDEX idx_system_settings_category ON system_settings(category);

-- Insert default settings
INSERT INTO system_settings (key, value, category, description) VALUES
    ('app_name', 'Figaró', 'general', 'Nombre de la aplicación'),
    ('app_version', '2.0.0', 'general', 'Versión de la aplicación'),
    ('default_timezone', 'Europe/Madrid', 'general', 'Zona horaria predeterminada'),
    ('default_language', 'es', 'general', 'Idioma predeterminado del sistema'),
    ('maintenance_mode', 'false', 'general', 'Modo mantenimiento activado'),
    
    ('session_timeout', '30', 'security', 'Tiempo de sesión en minutos'),
    ('max_login_attempts', '5', 'security', 'Máximo intentos de login'),
    ('require_uppercase', 'true', 'security', 'Requerir mayúsculas en contraseñas'),
    ('require_numbers', 'true', 'security', 'Requerir números en contraseñas'),
    ('require_special', 'false', 'security', 'Requerir caracteres especiales en contraseñas'),
    ('min_password_length', '8', 'security', 'Longitud mínima de contraseña'),
    
    ('smtp_host', '', 'email', 'Servidor SMTP'),
    ('smtp_port', '587', 'email', 'Puerto SMTP'),
    ('smtp_username', '', 'email', 'Usuario SMTP'),
    ('smtp_password', '', 'email', 'Contraseña SMTP'),
    ('smtp_from_email', '', 'email', 'Email remitente'),
    ('smtp_from_name', 'Figaró', 'email', 'Nombre remitente'),
    ('smtp_encryption', 'tls', 'email', 'Tipo de cifrado SMTP'),
    
    ('backup_enabled', 'true', 'backup', 'Respaldos automáticos habilitados'),
    ('backup_frequency', 'daily', 'backup', 'Frecuencia de respaldos'),
    ('backup_retention', '30', 'backup', 'Días de retención de respaldos'),
    ('backup_time', '02:00', 'backup', 'Hora de respaldo automático');