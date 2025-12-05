
import os
import json
from datetime import datetime
import plotly.graph_objects as go

def parse_log_files(log_dir):
    """
    Parses all access.log-* files in the given directory.
    """
    timestamps = []
    response_times = []
    
    for filename in os.listdir(log_dir):
        if filename.startswith("access.log-"):
            filepath = os.path.join(log_dir, filename)
            with open(filepath, 'r', encoding='utf-8') as f:
                for line in f:
                    try:
                        log_entry = json.loads(line)
                        
                        # Parse time_local
                        time_str = log_entry.get("time_local")
                        if time_str:
                            # The format is like "01/Nov/2025:03:35:52 +0800"
                            dt_object = datetime.strptime(time_str, '%d/%b/%Y:%H:%M:%S %z')
                            timestamps.append(dt_object)
                        
                        # Parse upstream_response_time
                        response_time_str = log_entry.get("upstream_response_time", "0")
                        
                        # Handle cases like "0.000, 0.000, 5.374" by taking the last value
                        if ',' in response_time_str:
                            response_time = float(response_time_str.split(',')[-1].strip())
                        else:
                            response_time = float(response_time_str)
                            
                        response_times.append(response_time)

                    except (json.JSONDecodeError, ValueError, IndexError) as e:
                        print(f"Skipping malformed line in {filename}: {line.strip()} - Error: {e}")

    return timestamps, response_times

def create_plots(timestamps, response_times, output_dir):
    """
    Creates scatter and line plots and saves them as HTML files.
    """
    if not timestamps or not response_times:
        print("No data to plot.")
        return

    # Create Scatter Plot
    scatter_fig = go.Figure(data=go.Scatter(
        x=timestamps,
        y=response_times,
        mode='markers',
        name='Upstream Response Time'
    ))
    scatter_fig.update_layout(
        title='Upstream Response Time (Scatter Plot)',
        xaxis_title='Time',
        yaxis_title='Response Time (s)'
    )
    scatter_plot_path = os.path.join(output_dir, 'scatter_plot.html')
    scatter_fig.write_html(scatter_plot_path)
    print(f"Scatter plot saved to {os.path.abspath(scatter_plot_path)}")

    # Create Line Plot
    # Sorting data by timestamp for a coherent line plot
    sorted_data = sorted(zip(timestamps, response_times))
    sorted_timestamps, sorted_response_times = zip(*sorted_data)

    line_fig = go.Figure(data=go.Scatter(
        x=sorted_timestamps,
        y=sorted_response_times,
        mode='lines+markers',
        name='Upstream Response Time'
    ))
    line_fig.update_layout(
        title='Upstream Response Time (Line Plot)',
        xaxis_title='Time',
        yaxis_title='Response Time (s)'
    )
    line_plot_path = os.path.join(output_dir, 'line_plot.html')
    line_fig.write_html(line_plot_path)
    print(f"Line plot saved to {os.path.abspath(line_plot_path)}")


if __name__ == "__main__":
    # The script is expected to be in the 'log' directory.
    log_directory = os.path.dirname(os.path.abspath(__file__))
    
    timestamps, response_times = parse_log_files(log_directory)
    create_plots(timestamps, response_times, log_directory)
