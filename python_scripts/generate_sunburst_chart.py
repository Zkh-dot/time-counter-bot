import json
import sys
import plotly.graph_objects as go

def build_hierarchical_data(nodes):
    """Создает иерархическую структуру данных для Sunburst диаграммы."""
    node_dict = {node["id"]: node for node in nodes}
    
    for node in nodes:
        node["children"] = []

    root_nodes = []
    
    for node in nodes:
        parent_id = node["parent_id"]
        if parent_id is not None and parent_id in node_dict:
            node_dict[parent_id]["children"].append(node)
        else:
            root_nodes.append(node)
    
    return root_nodes

def calculate_durations(node):
    """Рекурсивно вычисляет продолжительность для родительских узлов."""
    if node.get("duration") is None:
        node["duration"] = sum(calculate_durations(child) for child in node["children"])
    return node["duration"]

def generate_sunburst_chart(json_data, output_file):
    """Генерирует и сохраняет вложенную круговую диаграмму (sunburst chart)."""
    try:
        data = json.loads(json_data)
        root_nodes = build_hierarchical_data(data["nodes"])

        labels, parents, values, ids = [], [], [], []

        def traverse(node, parent_label=""):
            """Рекурсивно проходит по дереву и заполняет данные для графика."""
            label = node["name"] if not parent_label else f"{parent_label} / {node['name']}"
            labels.append(label)
            parents.append(parent_label if parent_label else "")
            values.append(node["duration"] if node["duration"] is not None else 0)
            ids.append(node["id"])

            for child in node["children"]:
                traverse(child, label)

        # Вычисляем суммы для родительских узлов
        for root in root_nodes:
            calculate_durations(root)
            traverse(root)

        fig = go.Figure(go.Sunburst(
            labels=labels,
            parents=parents,
            values=values,
            branchvalues="total",
            insidetextorientation="radial",
            marker=dict(colorscale="Blues"),
        ))

        fig.update_layout(
            title="Распределение времени по категориям",
            margin=dict(t=50, l=0, r=0, b=0)
        )

        fig.write_image(output_file, scale=2)
        print(f"✅ График сохранен в {output_file}")

    except Exception as e:
        print(f"❌ Ошибка генерации диаграммы: {e}")

if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Использование: python3 generate_sunburst_chart.py '<JSON>' <output_file>")
        sys.exit(1)

    json_data = sys.argv[1]
    output_file = sys.argv[2]
    generate_sunburst_chart(json_data, output_file)
