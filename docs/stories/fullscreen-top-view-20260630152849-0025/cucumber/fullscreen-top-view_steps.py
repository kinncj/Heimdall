# Step definitions for: Full-screen single-host top view (v2.2.0)
# Framework: behave  https://behave.readthedocs.io
# Each stub raises NotImplementedError until implemented.
from behave import given, when, then  # noqa: F401


# --- Keybind swap (p = processes, t = top) ---

@given(u"the dashboard is showing the fleet grid")
def step_the_dashboard_is_showing_the_fleet_grid(context):
    raise NotImplementedError(u"STEP: given the dashboard is showing the fleet grid")


@given(u"the dashboard has a focused host")
def step_the_dashboard_has_a_focused_host(context):
    raise NotImplementedError(u"STEP: given the dashboard has a focused host")


@given(u"the operator is in the full-screen top view")
def step_the_operator_is_in_the_full_screen_top_view(context):
    raise NotImplementedError(u"STEP: given the operator is in the full-screen top view")


@when(u'the operator presses "p"')
def step_the_operator_presses_p(context):
    raise NotImplementedError(u'STEP: when the operator presses "p"')


@when(u'the operator presses "t"')
def step_the_operator_presses_t(context):
    raise NotImplementedError(u'STEP: when the operator presses "t"')


@when(u'the operator presses "esc" or "q"')
def step_the_operator_presses_esc_or_q(context):
    raise NotImplementedError(u'STEP: when the operator presses "esc" or "q"')


@then(u'the process table opens labelled "processes"')
def step_the_process_table_opens_labelled_processes(context):
    raise NotImplementedError(u'STEP: then the process table opens labelled "processes"')


@then(u"the full-screen single-host top view opens for that host")
def step_the_full_screen_top_view_opens_for_that_host(context):
    raise NotImplementedError(u"STEP: then the full-screen single-host top view opens for that host")


@then(u"the dashboard returns to the fleet grid")
def step_the_dashboard_returns_to_the_fleet_grid(context):
    raise NotImplementedError(u"STEP: then the dashboard returns to the fleet grid")


# --- Panel content ---

@given(u"the operator opens the top view for a focused host")
def step_the_operator_opens_the_top_view_for_a_focused_host(context):
    raise NotImplementedError(u"STEP: given the operator opens the top view for a focused host")


@given(u"the focused host has recorded metric history")
def step_the_focused_host_has_recorded_metric_history(context):
    raise NotImplementedError(u"STEP: given the focused host has recorded metric history")


@given(u"the operator opens the top view")
def step_the_operator_opens_the_top_view(context):
    raise NotImplementedError(u"STEP: given the operator opens the top view")


@when(u"the top view renders its braille sparklines")
def step_the_top_view_renders_its_braille_sparklines(context):
    raise NotImplementedError(u"STEP: when the top view renders its braille sparklines")


@when(u"the accelerator panel renders")
def step_the_accelerator_panel_renders(context):
    raise NotImplementedError(u"STEP: when the accelerator panel renders")


@then(u"it shows per-core CPU bars, a CPU utilisation sparkline, and CPU frequency")
def step_it_shows_cpu_panel(context):
    raise NotImplementedError(u"STEP: then it shows per-core CPU bars, a CPU utilisation sparkline, and CPU frequency")


@then(u"it shows memory used and swap with a memory-bandwidth sparkline")
def step_it_shows_memory_panel(context):
    raise NotImplementedError(u"STEP: then it shows memory used and swap with a memory-bandwidth sparkline")


@then(u"it shows power for package, cpu, gpu, and npu with a power sparkline")
def step_it_shows_power_panel(context):
    raise NotImplementedError(u"STEP: then it shows power for package, cpu, gpu, and npu with a power sparkline")


@then(u"it shows GPU and NPU utilisation, VRAM, and temperature")
def step_it_shows_gpu_npu_panel(context):
    raise NotImplementedError(u"STEP: then it shows GPU and NPU utilisation, VRAM, and temperature")


@then(u"it shows network and disk sparklines")
def step_it_shows_network_and_disk_sparklines(context):
    raise NotImplementedError(u"STEP: then it shows network and disk sparklines")


@then(u"it shows the process list")
def step_it_shows_the_process_list(context):
    raise NotImplementedError(u"STEP: then it shows the process list")


@then(u"the sparklines are drawn from the existing per-host history buffers, not a new collector")
def step_sparklines_from_existing_history(context):
    raise NotImplementedError(u"STEP: then the sparklines are drawn from the existing per-host history buffers, not a new collector")


@then(u'it is labelled "NPU" rather than "ANE"')
def step_it_is_labelled_npu_rather_than_ane(context):
    raise NotImplementedError(u'STEP: then it is labelled "NPU" rather than "ANE"')


# --- Screen-size awareness ---

@given(u"a focused host shown in the top view")
def step_a_focused_host_shown_in_the_top_view(context):
    raise NotImplementedError(u"STEP: given a focused host shown in the top view")


@given(u"a focused host shown in the top view at any width tier")
def step_a_focused_host_shown_at_any_width_tier(context):
    raise NotImplementedError(u"STEP: given a focused host shown in the top view at any width tier")


@given(u"the top view content is taller than the terminal height")
def step_top_view_content_taller_than_terminal(context):
    raise NotImplementedError(u"STEP: given the top view content is taller than the terminal height")


@given(u"the operator is connected over SSH from Termius on a phone in portrait")
def step_operator_connected_from_termius_portrait(context):
    raise NotImplementedError(u"STEP: given the operator is connected over SSH from Termius on a phone in portrait")


@when(u"the terminal is at least 100 columns wide")
def step_terminal_at_least_100_columns(context):
    raise NotImplementedError(u"STEP: when the terminal is at least 100 columns wide")


@when(u"the terminal is between 60 and 99 columns wide")
def step_terminal_between_60_and_99_columns(context):
    raise NotImplementedError(u"STEP: when the terminal is between 60 and 99 columns wide")


@when(u"the terminal is between 40 and 59 columns wide")
def step_terminal_between_40_and_59_columns(context):
    raise NotImplementedError(u"STEP: when the terminal is between 40 and 59 columns wide")


@when(u"the terminal is narrower than 40 columns")
def step_terminal_narrower_than_40_columns(context):
    raise NotImplementedError(u"STEP: when the terminal is narrower than 40 columns")


@when(u"the top view renders")
def step_the_top_view_renders(context):
    raise NotImplementedError(u"STEP: when the top view renders")


@when(u"the operator scrolls with the up and down keys")
def step_the_operator_scrolls_with_up_and_down(context):
    raise NotImplementedError(u"STEP: when the operator scrolls with the up and down keys")


@when(u"the top view renders at a width narrower than 40 columns")
def step_top_view_renders_narrower_than_40(context):
    raise NotImplementedError(u"STEP: when the top view renders at a width narrower than 40 columns")


@then(u"panels render in a two-column grid with full sparklines and a multi-column per-core matrix")
def step_panels_two_column_grid(context):
    raise NotImplementedError(u"STEP: then panels render in a two-column grid with full sparklines and a multi-column per-core matrix")


@then(u"panels stack in a single column, sparklines are kept, and per-core bars wrap to fewer columns")
def step_panels_single_column_medium(context):
    raise NotImplementedError(u"STEP: then panels stack in a single column, sparklines are kept, and per-core bars wrap to fewer columns")


@then(u"panels stack, sparklines are shortened, and per-core collapses to an aggregate bar with a core-count summary")
def step_panels_narrow_aggregate(context):
    raise NotImplementedError(u"STEP: then panels stack, sparklines are shortened, and per-core collapses to an aggregate bar with a core-count summary")


@then(u"only key numbers (cpu%, mem%, power, temp) are shown, one value per line, with no graphs and no per-core bars")
def step_key_numbers_only(context):
    raise NotImplementedError(u"STEP: then only key numbers are shown, one value per line, with no graphs and no per-core bars")


@then(u"every rendered line fits within the terminal width")
def step_every_line_fits_terminal_width(context):
    raise NotImplementedError(u"STEP: then every rendered line fits within the terminal width")


@then(u"the header and footer stay fixed, the body scrolls, and the frame never exceeds the terminal height")
def step_header_footer_fixed_body_scrolls(context):
    raise NotImplementedError(u"STEP: then the header and footer stay fixed, the body scrolls, and the frame never exceeds the terminal height")


@then(u"it shows the key-numbers-only layout with no clipped lines")
def step_key_numbers_only_no_clipping(context):
    raise NotImplementedError(u"STEP: then it shows the key-numbers-only layout with no clipped lines")


# --- Graceful degradation (Scenario Outline) ---

@given(u'a "{platform}" host focused in the top view')
def step_a_platform_host_focused(context, platform):
    raise NotImplementedError(u'STEP: given a "%s" host focused in the top view' % platform)


@when(u'the top view renders "{metric}"')
def step_the_top_view_renders_metric(context, metric):
    raise NotImplementedError(u'STEP: when the top view renders "%s"' % metric)


@then(u'it shows "—" rather than an error or a fabricated 0')
def step_it_shows_dash_not_error_or_zero(context):
    raise NotImplementedError(u'STEP: then it shows "—" rather than an error or a fabricated 0')


@then(u"it shows a real value rather than a dash")
def step_it_shows_a_real_value(context):
    raise NotImplementedError(u"STEP: then it shows a real value rather than a dash")


# --- NPU rename, backward compatible ---

@given(u'a daemon that reports accelerator power under the legacy key "power.ane"')
def step_daemon_reports_legacy_power_ane(context):
    raise NotImplementedError(u'STEP: given a daemon that reports accelerator power under the legacy key "power.ane"')


@given(u'a daemon that reports accelerator power under "power.npu"')
def step_daemon_reports_power_npu(context):
    raise NotImplementedError(u'STEP: given a daemon that reports accelerator power under "power.npu"')


@when(u"the dashboard ingests that metric")
def step_the_dashboard_ingests_that_metric(context):
    raise NotImplementedError(u"STEP: when the dashboard ingests that metric")


@then(u'it is stored under the canonical key "power.npu"')
def step_stored_under_canonical_power_npu(context):
    raise NotImplementedError(u'STEP: then it is stored under the canonical key "power.npu"')


@then(u"the top view renders it under the NPU label")
def step_top_view_renders_under_npu_label(context):
    raise NotImplementedError(u"STEP: then the top view renders it under the NPU label")


@then(u"it is rendered under the NPU label without further translation")
def step_rendered_under_npu_no_translation(context):
    raise NotImplementedError(u"STEP: then it is rendered under the NPU label without further translation")


# --- Backward compatibility in a mixed fleet ---

@given(u"a mixed fleet where the focused host runs an older daemon without the new metric keys")
def step_mixed_fleet_older_daemon(context):
    raise NotImplementedError(u"STEP: given a mixed fleet where the focused host runs an older daemon without the new metric keys")


@given(u"a v2.2 daemon emitting the new metric keys")
def step_v23_daemon_emitting_new_keys(context):
    raise NotImplementedError(u"STEP: given a v2.2 daemon emitting the new metric keys")


@when(u"the operator opens the top view for that host")
def step_operator_opens_top_view_for_that_host(context):
    raise NotImplementedError(u"STEP: when the operator opens the top view for that host")


@when(u"an older hub or dashboard receives the stream")
def step_older_hub_or_dashboard_receives_stream(context):
    raise NotImplementedError(u"STEP: when an older hub or dashboard receives the stream")


@then(u'the panels render, the missing metrics show "—", and no error is raised')
def step_panels_render_missing_dash_no_error(context):
    raise NotImplementedError(u'STEP: then the panels render, the missing metrics show "—", and no error is raised')


@then(u"it ignores the unknown keys and keeps working")
def step_ignores_unknown_keys_keeps_working(context):
    raise NotImplementedError(u"STEP: then it ignores the unknown keys and keeps working")
