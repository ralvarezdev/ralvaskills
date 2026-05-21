---
name: ros2-architect
version: 1.0.0
description: ROS2 Kilted standards — workspace and package layout (colcon + ament_cmake/ament_python), lifecycle nodes, topics with explicit QoS, services, actions, parameters, Python launch DSL. Cross-platform development env via Pixi + RoboStack (replaces Ubuntu-only apt install). Both C++ (rclcpp) and Python (rclpy) first-class. Use when scaffolding or reviewing a ROS2 workspace, designing node/topic structure, or setting up reproducible robotics builds.
---

# ROS2 Architecture

Targets **ROS2 Kilted Kaiju** (May 2025). C++ (`rclcpp`) and Python (`rclpy`) treated as equal first-class targets. **Development environment via Pixi + RoboStack** — the modern, cross-platform alternative to the `sudo apt install ros-kilted-*` flow. See [STACK.md](STACK.md) for pinned tool versions.

## 1. Workspace layout

A ROS2 workspace is the root of one or more packages built together by `colcon`. Pixi sits at the workspace root and gives every contributor the same ROS distro, the same compiler, the same Python — on Linux, macOS, or Windows.

```
my_robot_ws/
├── pixi.toml                          # Pixi manifest: ROS distro + Python + tasks
├── pixi.lock                          # cross-platform lockfile (committed)
├── .gitignore                         # ignore build/ install/ log/ .pixi/
├── src/                               # all source packages
│   ├── my_robot_msgs/                 # custom interfaces (msg/srv/action)
│   │   ├── package.xml
│   │   ├── CMakeLists.txt
│   │   ├── msg/
│   │   ├── srv/
│   │   └── action/
│   ├── my_robot_drivers/              # C++ ament_cmake package
│   │   ├── package.xml
│   │   ├── CMakeLists.txt
│   │   ├── include/my_robot_drivers/
│   │   └── src/
│   ├── my_robot_navigation/           # Python ament_python package
│   │   ├── package.xml
│   │   ├── setup.py
│   │   ├── my_robot_navigation/
│   │   └── launch/
│   └── my_robot_bringup/              # launch + config only; no code
│       ├── package.xml
│       ├── launch/
│       └── config/
├── build/                             # colcon build outputs (gitignored)
├── install/                           # colcon install dir (gitignored)
└── log/                               # colcon logs (gitignored)
```

- **`src/` holds packages**, never code at the workspace root.
- **One package, one responsibility** — drivers, navigation, perception, custom messages, bringup (launch/config). Don't put everything in one `my_robot` package.
- **`_msgs` packages hold ONLY interfaces** (msg/srv/action). They get rebuilt rarely and consumed by everyone — keep code out.
- **`_bringup` package** holds launch files and config for assembling the full robot stack. No source code.

## 2. Pixi for the development environment

Pixi solves the historical ROS2 pain: needing a specific Ubuntu version, conflicting Python installs, broken updates from `apt`. RoboStack channel on conda-forge distributes ROS distros as conda packages, so Pixi can install ROS2 anywhere conda runs — Linux, macOS, Windows.

```toml
# pixi.toml
[project]
name = "my_robot_ws"
channels = ["https://prefix.dev/conda-forge", "https://prefix.dev/robostack-staging"]
platforms = ["linux-64", "linux-aarch64", "osx-arm64", "win-64"]
description = "My robot workspace"

[dependencies]
python                 = "3.12.*"
ros-kilted-desktop     = "*"
ros-kilted-rclcpp      = "*"
ros-kilted-rclpy       = "*"
ros-kilted-launch      = "*"
ros-kilted-launch-ros  = "*"
colcon-common-extensions = "*"
compilers              = "*"      # C++ toolchain (gcc / clang / MSVC)
cmake                  = "*"

[tasks]
build = "colcon build --symlink-install"
test  = "colcon test && colcon test-result --verbose"
clean = "rm -rf build/ install/ log/"
run   = { cmd = "ros2 launch my_robot_bringup robot.launch.py", depends-on = ["build"] }
```

```bash
pixi install                # one-time: install ROS + toolchain into .pixi/
pixi shell                  # enter the env
pixi run build              # build everything
pixi run run                # launch the robot
```

- **`pixi.lock` is committed.** Reproducibility is the whole point.
- **`platforms = [...]`** declares every OS/arch your team uses. CI uses the same lockfile.
- **No `sudo apt install` on contributor machines.** Pixi's `.pixi/` directory is isolated; the system stays clean.
- **CI uses Pixi too.** `pixi run build` in GitHub Actions = the same env every developer uses.
- **`colcon build --symlink-install`** in dev — Python changes hot-reload without rebuilding. Drop the flag for release builds.

## 3. Package structure — `ament_cmake` vs `ament_python`

The build type is declared in `package.xml`. Use `ament_cmake` for C++ (and for interface-only `_msgs` packages); `ament_python` for pure-Python packages.

```xml
<!-- ament_cmake — C++ packages and _msgs packages -->
<package format="3">
  <name>my_robot_drivers</name>
  <version>0.1.0</version>
  <description>Hardware drivers for My Robot</description>
  <maintainer email="me@example.com">Me</maintainer>
  <license>Apache-2.0</license>

  <buildtool_depend>ament_cmake</buildtool_depend>
  <depend>rclcpp</depend>
  <depend>std_msgs</depend>
  <depend>my_robot_msgs</depend>

  <test_depend>ament_lint_auto</test_depend>
  <test_depend>ament_lint_common</test_depend>

  <export><build_type>ament_cmake</build_type></export>
</package>
```

```xml
<!-- ament_python — Python packages -->
<package format="3">
  <name>my_robot_navigation</name>
  <version>0.1.0</version>
  <description>Navigation logic for My Robot</description>
  <maintainer email="me@example.com">Me</maintainer>
  <license>Apache-2.0</license>

  <exec_depend>rclpy</exec_depend>
  <exec_depend>my_robot_msgs</exec_depend>

  <test_depend>ament_copyright</test_depend>
  <test_depend>ament_flake8</test_depend>
  <test_depend>ament_pep257</test_depend>
  <test_depend>python3-pytest</test_depend>

  <export><build_type>ament_python</build_type></export>
</package>
```

- **`format="3"`** for `package.xml`. Older formats still work but lack features.
- **Every dep declared.** Implicit dependencies (`std_msgs` showing up because something else pulled it) is a latent bug.
- **`depend`** vs `build_depend` / `exec_depend`: prefer the unified `<depend>` tag when the dependency is needed at both build and runtime. Split only when they differ.
- **Tests declared as `test_depend`.** Linters (`ament_lint_*`) are part of every package's test suite.

## 4. Nodes — lifecycle, executors, single-responsibility

A **node** is the unit of computation. One process can host one or many nodes (`rclcpp::executors::MultiThreadedExecutor` / `rclpy.executors`).

- **One node, one job.** A node that publishes IMU data, listens for joystick input, *and* runs path planning has three responsibilities. Three nodes.
- **Lifecycle nodes** for anything with non-trivial init: `unconfigured → inactive → active → finalized`. Allows the launch system to bring up nodes in the right order, retry init, and gracefully shut down. Use for any node that holds an open connection (camera, motor controller, network socket).
- **Standard `Node`** for stateless or trivially-init nodes.
- **Single-threaded executor by default**; switch to multi-threaded only when callbacks genuinely contend (e.g. a long-running service callback shouldn't block fast topic callbacks). Multi-threaded executors require thread-safe callback groups — easy to get wrong.
- **Node names: lowercase snake_case**, namespaced by component (`/imu_driver`, `/path_planner`).
- **No business logic in `main.cpp` / `main.py`.** Construct the node, spin, shut down. Logic lives in node classes.

## 5. Topics & QoS — explicit profiles

Every publisher / subscriber declares an explicit QoS profile. Default behavior **bites**.

| QoS profile | When |
|---|---|
| **Sensor data** (`SensorDataQoS`) | High-rate sensor topics (camera, IMU, lidar) — best-effort reliability, keep-last depth, drop on overflow |
| **System default** | Slow, important data — reliable, keep-last 10 |
| **Parameters / services** (`ParametersQoS`) | Configuration, infrequent reliable delivery |
| **Custom** | When none of the above fit — declare every QoS attribute explicitly |

```cpp
// rclcpp — explicit QoS for a camera feed
auto qos = rclcpp::SensorDataQoS();   // best-effort, depth=5
auto pub = create_publisher<sensor_msgs::msg::Image>("/camera/image", qos);
```

```python
# rclpy — same idea
from rclpy.qos import qos_profile_sensor_data
self.pub = self.create_publisher(Image, '/camera/image', qos_profile_sensor_data)
```

- **Reliable + KEEP_ALL on high-rate topics is a memory leak.** A slow subscriber backs up the publisher's queue.
- **Best-effort + KEEP_LAST(N)** is right for "tell me the latest, drop the rest" — sensor streams.
- **Reliable + KEEP_LAST(10)** for low-rate state ("/robot/battery", "/system/status").
- **Topic names use plural / hierarchical paths**: `/sensors/imu/data`, `/control/cmd_vel`. Namespace by subsystem.
- **`rclcpp::Reliability::Reliable` between publisher and subscriber must match.** Mismatched QoS means subscribers silently fail to connect (visible only in `ros2 topic info -v`).

## 6. Services and actions

| | Service | Action |
|---|---|---|
| Purpose | Synchronous request/response | Long-running task with feedback + cancellation |
| Pattern | `Add(a, b) → c` | `MoveToGoal(pose) → feedback(progress) → result(success)` |
| Use when | Operation completes in < 100 ms | Operation takes seconds-to-minutes |
| Pitfall | Blocking callbacks freeze the executor | Action server complexity (goal/feedback/result/cancel) |

- **Services for short queries**: parameter get/set, mode change, calibration trigger.
- **Actions for everything navigation-like, manipulation-like, planning-like.**
- **Never call a service synchronously from inside a topic callback.** That's an executor deadlock. Use `async_send_request` (C++) / `call_async` (rclpy) and handle the future.

## 7. Parameters — declared, typed, validated

- **Declare every parameter on node init** with `declare_parameter(name, default, descriptor)`. Undeclared parameters fail to read.
- **Provide a `ParameterDescriptor`** with type and description. Visible in `ros2 param describe`.
- **Validate on change** via `add_on_set_parameters_callback` — reject bad values; never silently accept and crash later.
- **YAML config files** in the `_bringup` package, loaded by launch via `Node(parameters=[config_path])`. Per-robot, per-environment configs live there.
- **Use ranges and choices in the descriptor:** `IntegerRange`, `FloatingPointRange`, `additional_constraints` — the parameter system enforces them.

## 8. Launch files — Python DSL

Launch files describe how to bring up a system: which nodes to run, with which parameters, in which order, on which conditions.

```python
# launch/robot.launch.py
from launch import LaunchDescription
from launch.actions import DeclareLaunchArgument
from launch.substitutions import LaunchConfiguration
from launch_ros.actions import Node, LifecycleNode

def generate_launch_description():
    use_sim = LaunchConfiguration('use_sim')

    return LaunchDescription([
        DeclareLaunchArgument('use_sim', default_value='false'),

        LifecycleNode(
            package='my_robot_drivers',
            executable='camera_driver',
            name='camera_driver',
            namespace='sensors',
            parameters=[{'frame_rate': 30}],
            condition=...,
        ),

        Node(
            package='my_robot_navigation',
            executable='path_planner',
            name='path_planner',
            output='screen',
            parameters=['config/path_planner.yaml'],
        ),
    ])
```

- **Always Python launch (`*.launch.py`)** — XML launch is legacy.
- **Launch arguments for environment differences** (`use_sim`, `robot_model`). One launch file, many configurations.
- **Compose smaller launch files** with `IncludeLaunchDescription` — reuse the camera launch in both sim and real robot.
- **No business logic in launch files.** Launch files orchestrate; nodes compute.

## 9. Testing

- **Unit tests inside the package.** C++: GTest. Python: pytest.
- **`ament_lint_*`** in every package's test suite — `ament_cppcheck`, `ament_cpplint` (C++), `ament_flake8`, `ament_pep257` (Python). Failures fail CI.
- **Integration tests via `launch_testing`** — spin up a launch graph, assert nodes reach expected states, topics carry expected messages.
- **`ros2 doctor`** in CI before tests — catches missing deps, network issues, ROS_DOMAIN_ID conflicts.
- **Run tests in Pixi:** `pixi run test` → `colcon test && colcon test-result --verbose --all`.

## 10. Logging, debugging, deployment

- **`RCLCPP_INFO` / `self.get_logger().info()`** — ROS2 logger, integrates with the launch logger and external aggregators. Not `printf` / `print`.
- **Severity levels** (DEBUG / INFO / WARN / ERROR / FATAL) match the rest of our skills (e.g. [observability-architect §3](../../infra/observability-architect/SKILL.md#3-logging-structured-leveled-correlated)).
- **`ros2 topic echo`, `ros2 node info`, `ros2 service list`** are the field-debugger toolkit — make sure your nodes / topics / services are discoverable (correct namespace, QoS, lifecycle state).
- **`ros2 bag record / play`** for capturing real-robot data; integration tests replay against the same bag.
- **Containerized deploys:** Docker image built FROM a Pixi-installed base, or from `osrf/ros:kilted-desktop` if you need the upstream container. Multi-arch (amd64 + arm64) per [docker-architect §5](../../infra/docker-architect/SKILL.md#5-multi-arch-builds) for robot vs dev machine.

## 11. Cross-skill ties

- [docker-architect](../../infra/docker-architect/SKILL.md) — multi-arch builds for robot hardware (often arm64); the Pixi-in-Docker pattern for deployment.
- [observability-architect §3](../../infra/observability-architect/SKILL.md#3-logging-structured-leveled-correlated) — structured logging conventions; ROS2 logger forwards to aggregators.
- [grafana-architect](../../infra/grafana-architect/SKILL.md) — robot telemetry dashboards (CPU per node, topic latency, lifecycle states).
- [repo-tooling-architect](../../tooling/repo-tooling-architect/SKILL.md) — Pixi here partially overlaps with `mise`/`proto` (env management) and Task (task running). For ROS2 workspaces, Pixi covers both jobs; for polyglot repos with ROS2 + other services, mise + Pixi can coexist (Pixi owns the ROS2 environment).
- [python-architect](../../languages/python-architect/SKILL.md) / [go-architect](../../languages/go-architect/SKILL.md) — node implementations follow language conventions where they don't conflict with ROS2 patterns (callback structure, lifecycle).
