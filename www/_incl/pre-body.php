<?php require_once "utils.php";
if (isset($SKIP_AUTH) && $SKIP_AUTH === true) {
    // Skip auth because this is a log-in page
} else if (!is_logged_in()) {
    header("Location: /login.php");
    exit();
} else if (!isset($SKIP_CENTRO) && $SKIP_CENTRO != true) {
  if (!isset($_SESSION["centro"]) || !isset($_SESSION["aula"])) {
    header("Location: /elegir_centro.php");
    exit();
  }
}
?>
<!doctype html>
<html lang="es">

<head>
  <meta charset="utf-8" />
  <title>Figaró</title>
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <link href="/static/style.css" rel="stylesheet" />
</head>

<body id="top">
  <script>
    const showLoader = (message = "Solicitando...") => {
      const loader = document.querySelector("#loader");
      const loaderStat = document.querySelector("#loaderStat");
      if (loader) loader.style.display = "block";
      if (loaderStat) loaderStat.innerText = message;
    };
    
    const hideLoader = (message = "Descargando...") => {
      const loader = document.querySelector("#loader");
      const loaderStat = document.querySelector("#loaderStat");
      if (loader) loader.style.display = "none";
      if (loaderStat) loaderStat.innerText = message;
    };
    
    // Show "Solicitando..." if user reloads or leaves
    window.addEventListener("beforeunload", () => {
      showLoader("Solicitando...");
    });
    
    // Handle readyState (initial load)
    document.onreadystatechange = () => {
      if (document.readyState !== "complete") {
        showLoader("Descargando...");
      } else {
        hideLoader("Solicitando...");
      }
    };
    
    // Handle clicks on links and submits
    document.addEventListener("DOMContentLoaded", () => {
      document.querySelectorAll("a:not([target='_blank']):not([download])")
        .forEach(el => el.addEventListener("click", () => showLoader("Solicitando...")));
      
      document.querySelectorAll("form button[type='submit']")
        .forEach(btn => btn.addEventListener("click", () => showLoader("Solicitando...")));
    });
    
    // Handle back/forward navigation restores
    window.addEventListener("pageshow", event => {
      if (event.persisted) {
        hideLoader("Descargando...");
      }
    });
  </script>
  
  
  <center id="loader">
    <img loading="eager" src="/static/load.gif" width="200" height="200" />
    <h4 style="margin: 0;" id="loaderStat">Descargando...</h4>
    <progress style="width: calc(100% - 25px);"></progress>
  </center>
  <?php if (is_logged_in()) {
    $user = get_user_info();
  ?>
   <?php if (isset($_SESSION["centro"]) && isset($_SESSION["aula"])) { ?>
   <span style="font-family: monospace;"><?php echo $_SESSION["centro"]; ?> -> <?php echo $_SESSION["aula"]; ?></span>
   <?php } ?>
    <div style="background: lightcyan; border: 2px solid black; border-radius: 7px; padding: 5px;">
      <a class="button" href="/elegir_centro.php">
        <img loading="lazy" class="picto" src="/static/pictos/centro.png"><br>
        Elegir Centro
      </a>
      <?php if (user_has_access("materiales.index")) { ?>
        <a class="button" href="/materiales/index.php">
          <img loading="lazy" class="picto" src="/static/pictos/material_escolar.png"><br>
          Materiales
        </a>
      <?php } ?>
      <?php if (user_has_access("actividades.index")) { ?>
        <a class="button" href="/actividades/index.php">
          <img loading="lazy" class="picto" src="/static/pictos/actividades.png"><br>
          Actividades
        </a>
      <?php } ?>
      <?php if (user_has_access("archivos.index")) { ?>
        <a class="button" href="/notimpl.php">
          <img loading="lazy" class="picto" src="/static/pictos/archivos.png"><br>
          Archivos
        </a>
      <?php } ?>
      <?php if (user_has_access("ADMIN")) { ?>
        <a class="button" href="/admin/index.php">
          <img loading="lazy" class="picto" src="/static/pictos/datacenter.png"><br>
          Administración
        </a>
      <?php } ?>
      <a class="button" href="/logout.php">
        <img loading="lazy" class="picto" src="/static/pictos/candado.png"><br>
        Cerrar Sesión
      </a>
    </div>
  <?php } ?>
  <main id="container">
