<?php 
$SKIP_AUTH = true;
require_once "_incl/pre-body.php"; ?>
<center>
  <h2>¡Casi lo encuentras, <?php echo htmlspecialchars($user['display_name']); ?>!</h2>
</center>
<h3>Dado a que Figaró sigue en desarrollo, algunas funcionalidades pueden no estar disponibles.</h3>
<b>Lista de funcionalidades no implementadas:</b>
<ul>
  <li>Gestión de Actividades</li>
  <li>Calendario</li>
  <li>Email</li>
  <li>Gestión de Archivos</li>
  <li>Otras funcionalidades administrativas (próximamente)</li>
</ul>
<?php require_once "_incl/post-body.php"; ?>