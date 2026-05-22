# ROS2 Recipes

Reference scaffolds extracted from [SKILL.md](SKILL.md). Use these as templates when bootstrapping a workspace or a new package; the rules and *why* stay in the skill body.

## Workspace layout

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

- `src/` holds packages, never code at the workspace root.
- One package, one responsibility — drivers, navigation, perception, custom messages, bringup. Don't put everything in one `my_robot` package.
- `_msgs` packages hold ONLY interfaces (msg/srv/action). Rebuilt rarely, consumed by everyone — keep code out.
- `_bringup` packages hold launch files and config for assembling the full robot stack. No source code.

## QoS — explicit profiles

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

Profile reference:

| QoS profile | When |
|---|---|
| **Sensor data** (`SensorDataQoS`) | High-rate sensor topics (camera, IMU, lidar) — best-effort, keep-last, drop on overflow |
| **System default** | Slow, important data — reliable, keep-last 10 |
| **Parameters / services** (`ParametersQoS`) | Configuration, infrequent reliable delivery |
| **Custom** | None of the above fit — declare every QoS attribute explicitly |

- Reliable + KEEP_ALL on high-rate topics is a memory leak — a slow subscriber backs up the publisher's queue.
- Best-effort + KEEP_LAST(N) is right for "tell me the latest, drop the rest" — sensor streams.
- Reliable + KEEP_LAST(10) for low-rate state (`/robot/battery`, `/system/status`).
- Mismatched QoS between publisher and subscriber means subscribers silently fail to connect (visible only in `ros2 topic info -v`).

## `pixi.toml` — workspace manifest

Replace `<distro>` with `jazzy`, `kilted`, or `lyrical` consistently across the file. Pick once per workspace per [STACK.md](STACK.md#supported-distributions).

```toml
# pixi.toml
[project]
name = "my_robot_ws"
channels = ["https://prefix.dev/conda-forge", "https://prefix.dev/robostack-staging"]
platforms = ["linux-64", "linux-aarch64", "osx-arm64", "win-64"]
description = "My robot workspace"

[dependencies]
python                 = "3.12.*"
# Replace <distro> with `jazzy`, `kilted`, or `lyrical` (consistent across the file)
ros-<distro>-desktop     = "*"
ros-<distro>-rclcpp      = "*"
ros-<distro>-rclpy       = "*"
ros-<distro>-launch      = "*"
ros-<distro>-launch-ros  = "*"
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

## `package.xml` — `ament_cmake` (C++ and `_msgs`)

```xml
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

## `package.xml` — `ament_python`

```xml
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

## Launch file — Python DSL

Lifecycle node + standard node, with a launch argument, parameter dict, and YAML config. One launch, many configurations.

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
