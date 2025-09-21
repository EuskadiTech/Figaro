<!doctype html>
<html lang="es">

<head>
  <meta charset="utf-8" />
  <title>Figaró</title>
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <link href="/static/style.css" rel="stylesheet" />
  <link rel="icon" href="/static/logo.svg" sizes="any" type="image/svg+xml" />
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
    <img src="/static/load.gif" width="200" height="200" />
    <h4 style="margin: 0;" id="loaderStat">Descargando...</h4>
    <progress style="width: calc(100% - 25px);"></progress>
  </center>
  <details>
    <summary class="button">
      <img src="/pictos/mapa.png" height="70">
      <br>Navegación
    </summary>
    <div style="background: lightcyan; border: 2px solid black; border-radius: 7px; padding: 5px;">
      <a class="button">
        <img src="/pictos/aula.png" height="70">
        <br>Elegir Aula
      </a>
      <a class="button">
        <img src="/pictos/material_escolar.png" height="70">
        <br>Materiales
      </a>
      <a class="button">
        <img src="/pictos/actividades.png" height="70">
        <br>Actividades
      </a>
      <a class="button">
        <img src="/pictos/archivos.png" height="70">
        <br>Archivos
      </a>
    </div>
  </details>
  <main id="container">
