<?php
session_start();

// hide warnings and lower error reporting
error_reporting(E_ALL & ~E_WARNING & ~E_NOTICE);

$RUTA_DATOS = __DIR__ . "/../../data";
if (file_exists("/DATA")) {
    $RUTA_DATOS = "/DATA";
}

if (!file_exists($RUTA_DATOS)) {
    mkdir($RUTA_DATOS, 0755, true);
}

// User management functions using JSON files
function get_users_dir() {
    global $RUTA_DATOS;
    $users_dir = "$RUTA_DATOS/Usuarios";
    if (!file_exists($users_dir)) {
        mkdir($users_dir, 0755, true);
    }
    return $users_dir;
}

function load_user($username) {
    $users_dir = get_users_dir();
    $user_file = "$users_dir/$username.json";
    if (file_exists($user_file)) {
        return json_decode(file_get_contents($user_file), true);
    }
    return null;
}

function save_user($username, $user_data) {
    $users_dir = get_users_dir();
    $user_file = "$users_dir/$username.json";
    return file_put_contents($user_file, json_encode($user_data, JSON_PRETTY_PRINT)) !== false;
}

function get_all_users() {
    $users_dir = get_users_dir();
    $users = [];
    if (is_dir($users_dir)) {
        $files = glob("$users_dir/*.json");
        foreach ($files as $file) {
            $username = basename($file, '.json');
            $user_data = json_decode(file_get_contents($file), true);
            if ($user_data) {
                $users[$username] = $user_data;
            }
        }
    }
    return $users;
}

function delete_user($username) {
    $users_dir = get_users_dir();
    $user_file = "$users_dir/$username.json";
    if (file_exists($user_file)) {
        return unlink($user_file);
    }
    return false;
}

// Initialize demo user if no users exist
function ensure_demo_user() {
    $users = get_all_users();
    if (empty($users)) {
        $demo_user = [
            'password' => password_hash('demo', PASSWORD_DEFAULT),
            'auth' => ['ADMIN', 'module2'],
            'display_name' => 'Demo User',
            'email' => 'demo@example.com',
            'created_at' => date('c')
        ];
        save_user('demo', $demo_user);
    }
}

// Ensure demo user exists on every request
ensure_demo_user();
function login($username, $password)
{
    $user = load_user($username);
    if ($user && password_verify($password, $user['password'])) {
        setcookie("username", $username, time() + 3600, "/");
        setcookie("password", base64_encode($password), time() + 3600, "/");
        setcookie("loggedin", "yes", time() + 3600, "/");
        return true;
    }
    return false;
}
function logout()
{
    setcookie("username", "", time() - 3600, "/");
    setcookie("password", "", time() - 3600, "/");
}
function is_logged_in()
{
    // Undefined index fix
    if (!isset($_COOKIE["loggedin"]) || !isset($_COOKIE["username"]) || !isset($_COOKIE["password"])) {
        return false;
    }
    $user = load_user($_COOKIE["username"]);
    return $_COOKIE["loggedin"] == "yes" && $user && password_verify(base64_decode($_COOKIE["password"]), $user['password']);
}
function get_user_info()
{
    if (is_logged_in()) {
        return load_user($_COOKIE["username"]);
    }
    return null;
}
function user_has_access($module)
{
    if (is_logged_in()) {
        $user = load_user($_COOKIE["username"]);
        if (!$user) return false;
        
        // if user has access to ADMIN, allow all modules
        if (in_array('ADMIN', $user['auth'])) {
            return true;
        }
        // else check if user has access to the requested module
        return in_array($module, $user['auth']);
    }
    return false;
}

function require_permission($module)
{
    if (!user_has_access($module)) {
        header("Location: /index.php?flash=No+tienes+permiso+para+acceder+a+esta+página");
        exit();
    }
}

function set_selected_centro_aula($centro, $aula)
{
    $_SESSION['centro'] = $centro;
    $_SESSION['aula'] = $aula;
}

function get_selected_centro()
{
    return isset($_SESSION['centro']) ? $_SESSION['centro'] : null;
}

function get_selected_aula()
{
    return isset($_SESSION['aula']) ? $_SESSION['aula'] : null;
}

function is_centro_aula_selected()
{
    return isset($_SESSION['centro']) && isset($_SESSION['aula']);
}


function login_with_qr($qr_data)
{
    $parts = explode(':', $qr_data);
    if (count($parts) !== 3) {
        return false;
    }

    $username = $parts[0];
    $password = base64_decode($parts[1]);
    $hash = $parts[2];

    $user = load_user($username);
    if (!$user) {
        return false;
    }

    $expected_hash = hash('sha256', $username . ':' . $password);
    if (!hash_equals($expected_hash, $hash)) {
        return false;
    }

    return login($username, $password);
}

function get_centros() {
    global $RUTA_DATOS;
    $centros_path = "$RUTA_DATOS/Centros";
    if (!file_exists($centros_path)) {
        mkdir($centros_path, 0755, true);
    }
    $centros = array_filter(scandir($centros_path), function ($item) use ($centros_path) {
        return $item !== '.' && $item !== '..' && is_dir("$centros_path/$item");
    });
    return $centros;
}

function iso_to_es($iso_date) {
    $date = explode('T', $iso_date)[0];
    $parts = explode('-', $date);
    if (count($parts) !== 3) {
        return $iso_date; // return original if format is unexpected
    }
    // Invert array (YYYY-MM-DD to DD-MM-YYYY)
    return implode('/', array_reverse($parts));
}

function get_centro_config($centro_name) {
    global $RUTA_DATOS;
    $config_file = "$RUTA_DATOS/Centros/$centro_name/config.json";
    if (file_exists($config_file)) {
        return json_decode(file_get_contents($config_file), true);
    }
    return null;
}

function is_off_hours($activity_start, $activity_end, $centro_name) {
    $config = get_centro_config($centro_name);
    if (!$config || !isset($config['working_hours'])) {
        return false; // No config, assume always open
    }
    
    $start_datetime = new DateTime($activity_start);
    $end_datetime = new DateTime($activity_end);
    
    // Get day of week (lowercase)
    $day_name = strtolower($start_datetime->format('l'));
    
    // Check if the day is defined in working hours
    if (!isset($config['working_hours'][$day_name]) || $config['working_hours'][$day_name] === null) {
        return true; // Day is closed
    }
    
    $working_hours = $config['working_hours'][$day_name];
    $work_start = $working_hours['start'];
    $work_end = $working_hours['end'];
    
    // Convert times to comparable format
    $activity_start_time = $start_datetime->format('H:i');
    $activity_end_time = $end_datetime->format('H:i');
    
    // Check if activity is outside working hours
    if ($activity_start_time < $work_start || $activity_end_time > $work_end) {
        return true;
    }
    
    // If activity spans multiple days, check the end day too
    if ($start_datetime->format('Y-m-d') !== $end_datetime->format('Y-m-d')) {
        $end_day_name = strtolower($end_datetime->format('l'));
        if (!isset($config['working_hours'][$end_day_name]) || $config['working_hours'][$end_day_name] === null) {
            return true; // End day is closed
        }
    }
    
    return false;
}