<?php
require_once "../_incl/pre-body.php";

// Get activities
$centro = $_SESSION['centro'];
$aula = $_SESSION['aula'];

function get_activities_from_dir($dir, $is_global = false)
{
    $activities = [];
    if (!is_dir($dir)) {
        return $activities;
    }
    $files = array_diff(scandir($dir), ['.', '..']);
    foreach ($files as $file) {
        if (pathinfo($file, PATHINFO_EXTENSION) === 'json') {
            $content = file_get_contents("$dir/$file");
            $activity = json_decode($content, true);
            $activity["_global"] = $is_global;
            if ($activity) {
                $activities[] = $activity;
            }
        }
    }
    return $activities;
}

$local_activities = get_activities_from_dir("../../data/Actividades/$centro/$aula");
$global_activities = get_activities_from_dir("../../data/Actividades/_Global", true);
$all_activities = array_merge($local_activities, $global_activities);

// Filter for upcoming activities
$today = new DateTime();
$today->setTime(0, 0, 0); // Set time to beginning of the day for comparison

$upcoming_activities = array_filter($all_activities, function ($activity) use ($today) {
    $start_date = new DateTime($activity['start']);
    return $start_date >= $today;
});

// Sort activities by start date
usort($upcoming_activities, function ($a, $b) {
    return strtotime($a['start']) - strtotime($b['start']);
});

$date_formatter = new IntlDateFormatter(
    'es_ES',
    IntlDateFormatter::FULL,
    IntlDateFormatter::SHORT,
    null,
    null,
    'eeee, d \'de\' MMMM \'de\' yyyy HH:mm'
);
?>

<h2>Próximas Actividades</h2>

<a href="crear_actividad.php" class="btn btn-success" style="margin-bottom: 20px;">Crear nueva actividad</a>
<br>
<?php if (empty($upcoming_activities)): ?>
    <p>No hay próximas actividades.</p>
<?php else: ?>
    <?php foreach ($upcoming_activities as $activity): ?>
        <div
            style="background-color: lightcyan; padding: 10px; margin-top: 5px; border-radius: 25px; border: 2px solid black; display: inline-block;">
            <h3>
                <?php echo htmlspecialchars($activity['title']); ?>
            </h3>
            <p><img loading="lazy" class="picto" src="/static/pictos/dia.png">
                <span style="font-weight: bold; vertical-align: top; display: inline-block;"> Dia: <br>
                    <span style="font-size: xx-large;"><?php echo htmlspecialchars(iso_to_es($activity['start'])); ?>
                    </span> </span>
            </p>
            <p><img loading="lazy" class="picto" src="/static/pictos/inicio.png">
                <span style="font-weight: bold; vertical-align: top; display: inline-block;"> Inicio: <br>
                    <span style="font-size: xx-large;"><?php echo htmlspecialchars(explode("T", $activity['start'])[1]); ?>
                    </span> </span>
            </p>
            <p><img loading="lazy" class="picto" src="/static/pictos/fin.png">
                <span style="font-weight: bold; vertical-align: top; display: inline-block;"> Fin: <br>
                    <span style="font-size: xx-large;"><?php echo htmlspecialchars(explode("T", $activity['end'])[1]); ?>
                    </span> </span>
            </p>
            <p><b>Descripción:</b> <?php echo nl2br(htmlspecialchars($activity['description'])); ?></p>

            <a href="editar_actividad.php?id=<?php echo urlencode($activity['id']); ?>&global=<?php echo $activity['_global'] ? '1' : '0'; ?>"
                class="button">

                <img loading="lazy" class="picto" src="/static/pictos/escribir.png"><br>Editar
            </a>
            <a href="eliminar_actividad.php?id=<?php echo urlencode($activity['id']); ?>&global=<?php echo $activity['_global'] ? '1' : '0'; ?>"
                onclick="return confirm('¿Estás seguro?');" class="button rojo">
                <img loading="lazy" class="picto" src="/static/pictos/borrar.png"><br>Borrar</a>
        </div>
    <?php endforeach; ?>
<?php endif; ?>

<?php
require_once "../_incl/post-body.php";
?>