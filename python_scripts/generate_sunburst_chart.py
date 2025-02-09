import sys
import json
from typing import Any, Dict, List, Optional

import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns


# 🎨 Настройки графика
sns.set_style("whitegrid")
plt.rcParams.update({
    "font.size": 12,
    "axes.titlesize": 16
})


class Node:
    """Класс для представления узла дерева активностей."""

    def __init__(
        self, node_id: int, name: str, parent_id: Optional[int],
        duration: Optional[float]
    ) -> None:
        self.node_id = node_id
        self.name = name
        self.parent_id = parent_id
        self.duration = duration if duration is not None else 0.0
        self.children: List["Node"] = []


def build_tree(df: pd.DataFrame) -> List[Node]:
    """
    Преобразует DataFrame в древовидную структуру
    и вычисляет длительность у родительских узлов.
    """
    node_dict: Dict[int, Node] = {}

    # 🔹 Создаём узлы
    for _, row in df.iterrows():
        node = Node(
            node_id=row["id"],
            name=row["name"],
            parent_id=int(row["parent_id"]) if pd.notna(row["parent_id"]) else None,
            duration=row["duration"],
        )
        node_dict[node.node_id] = node

    # 🔹 Добавляем детей к родителям
    roots: List[Node] = []
    for node in node_dict.values():
        if node.parent_id is not None and node.parent_id in node_dict:
            node_dict[node.parent_id].children.append(node)
        else:
            roots.append(node)

    # 🔹 Рекурсивно вычисляем `duration` для родительских узлов
    def calculate_duration(node: Node) -> float:
        if node.children:
            node.duration = sum(
                calculate_duration(child) for child in node.children
            )
        return node.duration

    for root in roots:
        calculate_duration(root)

    return roots


def flatten_tree(node: Node, base_name: str = "") -> List[Dict[str, Any]]:
    """Разворачивает дерево в плоский список."""
    full_name = f"{base_name} / {node.name}" if base_name else node.name
    data = [{"name": full_name, "duration": node.duration}]
    for child in node.children:
        data.extend(flatten_tree(child, full_name))
    return data


def generate_sunburst_chart(json_data: str, output_file: str) -> None:
    """
    Создаёт круговую диаграмму (Sunburst) на основе JSON-данных.
    """
    try:
        # ✅ Загружаем данные и проверяем их
        data = json.loads(json_data)
        if "nodes" not in data:
            raise KeyError("JSON data must contain 'nodes' key")

        df = pd.DataFrame(data["nodes"])

        # ✅ Проверяем NaN в `duration`
        df["duration"] = df["duration"].fillna(0).astype(float)
        df["parent_id"] = df["parent_id"].apply(
            lambda x: int(x) if pd.notna(x) else None
        )

        # ✅ Строим дерево и разворачиваем в плоскую структуру
        tree_data = build_tree(df)
        flat_data = []
        for root in tree_data:
            flat_data.extend(flatten_tree(root))

        df_flat = pd.DataFrame(flat_data)

        # ✅ Проверяем, есть ли данные в `duration`
        if df_flat["duration"].sum() == 0:
            raise ValueError("Total duration is zero. Check your input data!")

        # 🎨 Настройки цветов
        colors = sns.color_palette("coolwarm", len(df_flat))

        # 🎨 Создаём круговую диаграмму (Sunburst)
        fig, ax = plt.subplots(
            figsize=(8, 8), dpi=150, subplot_kw=dict(polar=True)
        )
        total = df_flat["duration"].sum()
        start_angle = 0.0

        for i, row in df_flat.iterrows():
            angle = (row["duration"] / total) * 2 * 3.1416
            ax.barh(
                y=i, width=angle, left=start_angle, height=1,
                color=colors[i], label=row["name"]
            )
            start_angle += angle

        # 🎨 Легенда
        ax.legend(
            loc="center left", bbox_to_anchor=(1, 0.5), title="Активности"
        )

        # 🎨 Убираем оси
        ax.set_yticks([])
        ax.set_xticks([])
        ax.set_title(
            "Распределение времени по категориям",
            fontsize=16, fontweight="bold", pad=20
        )

        # 🎨 Сохраняем изображение
        plt.savefig(
            output_file, format="png",
            transparent=False, bbox_inches="tight"
        )
        print(f"✅ Chart saved as {output_file}")

    except Exception as error:
        print(f"❌ Error generating chart: {error}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print(
            "Usage: python3 generate_sunburst_chart.py '<json_data>' <output_file>",
            file=sys.stderr
        )
        sys.exit(1)

    input_json = sys.argv[1]
    output_path = sys.argv[2]

    generate_sunburst_chart(input_json, output_path)
