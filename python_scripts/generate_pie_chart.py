import sys
import json
import matplotlib.pyplot as plt


def generate_pie_chart(json_data, output_file):
    """
    Генерирует круговую диаграмму по JSON-данным и сохраняет её в файл.
    
    :param json_data: JSON-строка с данными (словарь: {название: значение}).
    :param output_file: Путь к файлу, куда сохранить изображение.
    """
    try:
        data = json.loads(json_data)
        labels = list(data.keys())
        values = list(data.values())

        plt.figure(figsize=(8, 8))
        plt.pie(values, labels=labels, autopct='%1.1f%%', startangle=140, colors=plt.cm.Paired.colors)
        plt.axis('equal')  # Сделать круговую диаграмму кругом

        plt.savefig(output_file, format='png')
        print(f"Chart saved as {output_file}")

    except Exception as e:
        print(f"Error generating chart: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: python3 generate_pie_chart.py '<json_data>' <output_file>", file=sys.stderr)
        sys.exit(1)

    json_data = sys.argv[1]
    output_file = sys.argv[2]

    generate_pie_chart(json_data, output_file)
