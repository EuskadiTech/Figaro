<?php
require_once "_incl/utils.php";
$SKIP_AUTH = false; // Make sure user is logged in to access this page
$SKIP_CENTRO = true; // Skip centro/aula check for this page

// Should we require permission to select centro/aula?
// No! because if user does not have a preassigned centro/aula, they need to be able to select one.
// require_permission('select_centro_aula');

// Handle form submission
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_POST['centro']) && isset($_POST['aula'])) {
    set_selected_centro_aula($_POST['centro'], $_POST['aula']);
    header("Location: /index.php"); // Redirect to home page after selection
    exit();
}

require_once "_incl/pre-body.php";

$centros_path = $RUTA_DATOS . '/Centros';
$centros = is_dir($centros_path) ? array_filter(scandir($centros_path), function ($item) use ($centros_path) {
    return is_dir($centros_path . '/' . $item) && !in_array($item, ['.', '..']);
}) : [];

$selected_centro = isset($_GET['centro']) ? $_GET['centro'] : null;
$aulas = [];
if ($selected_centro && in_array($selected_centro, $centros)) {
    $aulas_path = $centros_path . '/' . $selected_centro . '/Aulas';
    $aulas = is_dir($aulas_path) ? array_filter(scandir($aulas_path), function ($item) use ($aulas_path) {
        return is_dir($aulas_path . '/' . $item) && !in_array($item, ['.', '..']);
    }) : [];
}
?>


<?php if ($selected_centro): ?>
    <h2>Centro seleccionado: <strong><?php echo htmlspecialchars($selected_centro); ?></strong></h2>

    <h1>Elige un aula</h1>
    <form method="POST" action="elegir_centro.php">
        <input type="hidden" name="centro" value="<?php echo htmlspecialchars($selected_centro); ?>">
        <?php foreach ($aulas as $aula): ?>
            <button type="submit" name="aula" value="<?php echo htmlspecialchars($aula); ?>" class="button">
                <?php echo htmlspecialchars($aula); ?>
            </button>
        <?php endforeach; ?>
    </form>
<?php else: ?>
    <h1>Elige un centro</h1>
    <?php foreach ($centros as $centro): ?>
        <a href="?centro=<?php echo urlencode($centro); ?>" class="button">
            <?php echo htmlspecialchars($centro); ?>
        </a>
    <?php endforeach; ?>
<?php endif; ?>
</div>

<?php require_once "_incl/post-body.php"; ?>