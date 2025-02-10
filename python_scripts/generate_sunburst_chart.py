#!/usr/bin/env python
"""
Генерация sunburst-диаграммы для дерева активностей пользователя.

На вход подается JSON с узлами (activities), где каждый узел имеет:
- id: уникальный идентификатор,
- parent_id: идентификатор родителя (или null для корневых),
- name: название активности,
- duration: длительность (если узел-лист) или null (если промежуточный узел).

Диаграмма сохраняется в PNG-файл по указанному пути.
"""

import json
import sys
from collections import defaultdict

import matplotlib.pyplot as plt
import matplotlib.patches as mpatches
import numpy as np
import seaborn as sns


def generate_sunburst_chart(data, output_file):
    """
    Генерирует sunburst-диаграмму по данным о дереве активностей и сохраняет её
    в output_file.

    Для каждого узла отображаются название, суммарная длительность и процент от
    общего времени. Легенда содержит только листовые узлы (с duration), отсортированные
    по убыванию длительности.
    """
    # Строим словарь узлов и отображение: родитель -> список детей.
    nodes = {node["id"]: node for node in data["nodes"]}
    children = defaultdict(list)
    for node in data["nodes"]:
        if node["parent_id"] is not None:
            children[node["parent_id"]].append(node["id"])

    # Определяем корневые узлы (без parent_id).
    root_ids = [node_id for node_id, node in nodes.items()
                if node["parent_id"] is None]

    # Рекурсивное вычисление "общей длительности" для каждого узла.
    computed_duration = {}

    def compute_duration(node_id):
        node = nodes[node_id]
        if node.get("duration") is not None:
            d = node["duration"]
        else:
            d = sum(
                compute_duration(child_id)
                for child_id in children[node_id]
            )
        computed_duration[node_id] = d
        return d

    for node_id in root_ids:
        compute_duration(node_id)
    total_duration = sum(
        computed_duration[node_id] for node_id in root_ids
    )

    # Определяем максимальную глубину дерева.
    max_depth = 0

    def find_max_depth(node_id, depth):
        nonlocal max_depth
        if depth > max_depth:
            max_depth = depth
        for child_id in children[node_id]:
            find_max_depth(child_id, depth + 1)

    for node_id in root_ids:
        find_max_depth(node_id, 1)

    # Назначаем уникальные цвета для узлов (пастельная палитра).
    all_node_ids = list(nodes.keys())
    palette = sns.color_palette("pastel", len(all_node_ids))
    color_dict = {node_id: palette[i]
                  for i, node_id in enumerate(all_node_ids)}

    # Создаем фигуру и ось.
    fig, ax = plt.subplots(figsize=(8, 8))
    ax.set_aspect("equal")
    ax.axis("off")

    # Параметр: радиус пустого центра.
    base_radius = 1.0

    # Рисуем пустой белый круг в центре.
    center_circle = plt.Circle((0, 0), base_radius, color="white", zorder=10)
    ax.add_patch(center_circle)

    def draw_node(node_id, start_angle, angle_width, level):
        """
        Рекурсивно рисует сектор кольца для узла и его потомков.
        """
        inner_radius = base_radius + (level - 1)
        outer_radius = base_radius + level

        wedge = mpatches.Wedge(
            center=(0, 0),
            r=outer_radius,
            theta1=start_angle,
            theta2=start_angle + angle_width,
            width=outer_radius - inner_radius,
            facecolor=color_dict[node_id],
            edgecolor="white",
        )
        ax.add_patch(wedge)

        node_duration = computed_duration[node_id]
        node_percentage = node_duration / total_duration * 100
        display_text = (
            f"{nodes[node_id]['name']}\n"
            f"{node_duration} ({node_percentage:.1f}%)"
        )

        mid_angle = start_angle + angle_width / 2
        r_text = (inner_radius + outer_radius) / 2
        x_text = r_text * np.cos(np.deg2rad(mid_angle))
        y_text = r_text * np.sin(np.deg2rad(mid_angle))
        rotation = mid_angle - 90
        if rotation > 90:
            rotation -= 180
        if rotation < -90:
            rotation += 180

        ax.text(
            x_text,
            y_text,
            display_text,
            ha="center",
            va="center",
            fontsize=9,
            rotation=rotation,
            rotation_mode="anchor",
        )

        if children[node_id]:
            child_start_angle = start_angle
            parent_duration = computed_duration[node_id]
            for child_id in children[node_id]:
                child_duration = computed_duration[child_id]
                child_angle_width = angle_width * (child_duration / parent_duration)
                draw_node(
                    child_id,
                    child_start_angle,
                    child_angle_width,
                    level + 1,
                )
                child_start_angle += child_angle_width

    current_angle = 0
    for node_id in root_ids:
        node_angle_width = 360 * (
            computed_duration[node_id] / total_duration
        )
        draw_node(node_id, current_angle, node_angle_width, level=1)
        current_angle += node_angle_width

    # Формирование легенды для листовых узлов.
    legend_handles = []
    leaf_nodes = []
    for node_id, node in nodes.items():
        if node.get("duration") is not None:
            d = computed_duration[node_id]
            leaf_nodes.append((node_id, d))
    leaf_nodes.sort(key=lambda x: x[1], reverse=True)
    for node_id, d in leaf_nodes:
        p = d / total_duration * 100
        label = f"{nodes[node_id]['name']} - {d} ({p:.1f}%)"
        patch = mpatches.Patch(color=color_dict[node_id], label=label)
        legend_handles.append(patch)
    ax.legend(
        handles=legend_handles,
        loc="center left",
        bbox_to_anchor=(1, 0.5)
    )

    max_radius_value = base_radius + max_depth
    margin = 0.2 * max_radius_value
    ax.set_xlim(-max_radius_value - margin,
                max_radius_value + margin)
    ax.set_ylim(-max_radius_value - margin,
                max_radius_value + margin)

    plt.savefig(output_file, dpi=300, bbox_inches="tight")
    plt.close()


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Использование: python script.py '<json_data>' output.png")
        sys.exit(1)
    input_json = sys.argv[1]
    output_filename = sys.argv[2]
    try:
        data = json.loads(input_json)
        generate_sunburst_chart(data, output_filename)
        print(f"✅ Диаграмма успешно сохранена в {output_filename}")
    except Exception as e:
        print(f"❌ Ошибка генерации диаграммы: {e}")
