import sys
import json
import matplotlib.pyplot as plt
import seaborn as sns

# Настраиваем глобальный стиль seaborn
sns.set_style("whitegrid")
plt.rcParams.update({
    "font.size": 14,  # Увеличенный размер шрифта
    "axes.titlesize": 18,
    "axes.labelsize": 14,
    "xtick.labelsize": 12,
    "ytick.labelsize": 12,
    "legend.fontsize": 12
})


def generate_pie_chart(json_data, output_file):
    """
    Генерирует улучшенную круговую диаграмму с тенями и градиентами.
    
    :param json_data: JSON-строка с данными (словарь: {название: значение}).
    :param output_file: Путь к файлу, куда сохранить изображение.
    """
    try:
        # Декодируем JSON
        data = json.loads(json_data)
        labels = list(data.keys())
        values = list(data.values())

        # Создаём красивую цветовую палитру
        colors = sns.color_palette("pastel", len(labels))

        # Создаём фигуру и оси
        fig, ax = plt.subplots(figsize=(8, 8), dpi=150)

        # Рисуем круговую диаграмму с тенью
        wedges, texts, autotexts = ax.pie(
            values, labels=labels, autopct='%1.1f%%', startangle=140, 
            colors=colors, wedgeprops={"edgecolor": "gray", "linewidth": 1, "antialiased": True}
        )

        # Делаем тени для каждого сектора
        for w in wedges:
            w.set_path_effects([
                plt.matplotlib.patheffects.SimpleLineShadow(),
                plt.matplotlib.patheffects.Normal()
            ])

        # Настраиваем текст внутри кругов
        for text in texts + autotexts:
            text.set_fontsize(12)
            text.set_color("black")
            text.set_path_effects([
                plt.matplotlib.patheffects.withStroke(linewidth=3, foreground="white")
            ])

        # Делаем легенду
        ax.legend(wedges, labels, title="Категории", loc="best", frameon=True, shadow=True)

        # Делаем фон прозрачным
        fig.patch.set_alpha(0)

        # Сохраняем изображение
        plt.savefig(output_file, format="png", transparent=True)
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
