<?php require_once "_incl/pre-body.php"; ?>

<?php if (isset($_GET['flash'])): ?>
<div style="background-color: #ffebee; border: 2px solid #f44336; color: #c62828; padding: 15px; margin: 20px 0; border-radius: 10px; text-align: center;">
    <strong>⚠️ Acceso Denegado:</strong> <?php echo htmlspecialchars($_GET['flash']); ?>
</div>
<?php endif; ?>

<center>
  <h2>¡Hola <?php echo htmlspecialchars($user['display_name']); ?>!</h2>
  <h1>Bienvenidx a Figaró</h1>
  <em>Utiliza el menú superior para acceder a los modulos a los que tienes acceso</em>
</center>
<?php require_once "_incl/post-body.php"; ?>