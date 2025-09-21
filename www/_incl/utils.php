<?php
session_start();

$RUTA_DATOS = __DIR__ . "/../../data";
if (file_exists("/DATA")) {
    $RUTA_DATOS = "/DATA";
}

if (!file_exists($RUTA_DATOS)) {
    mkdir($RUTA_DATOS, 0755, true);
}

$users = [
    'demo' => [
        'password' => password_hash('demo', PASSWORD_DEFAULT),
        'auth' => ['ADMIN', 'module2'],
        "display_name" => "Demo User",
        'email' => 'demo@example.com',
    ],
    // Otros usuarios pueden ser añadidos aquí
];
function login($username, $password)
{
    global $users;
    if (isset($users[$username]) && password_verify($password, $users[$username]['password'])) {
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
    global $users;
    // Undefined index fix
    if (!isset($_COOKIE["loggedin"]) || !isset($_COOKIE["username"]) || !isset($_COOKIE["password"])) {
        return false;
    }
    return $_COOKIE["loggedin"] == "yes" && password_verify(base64_decode($_COOKIE["password"]), $users[$_COOKIE["username"]]['password']);
}
function get_user_info()
{
    global $users;
    if (is_logged_in()) {
        return $users[$_COOKIE["username"]];
    }
    return null;
}
function user_has_access($module)
{
    global $users;
    if (is_logged_in()) {
        // if user has access to ADMIN, allow all modules
        if (in_array('ADMIN', $users[$_COOKIE["username"]]['auth'])) {
            return true;
        }
        // else check if user has access to the requested module
        return in_array($module, $users[$_COOKIE["username"]]['auth']);
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
    global $users;
    $parts = explode(':', $qr_data);
    if (count($parts) !== 3) {
        return false;
    }

    $username = $parts[0];
    $password = base64_decode($parts[1]);
    $hash = $parts[2];

    if (!isset($users[$username])) {
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