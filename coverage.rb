#!ruby
require 'json'

$colors = {
  'red'         => '#e05d44',
  'orange'      => '#fe7d37',
  'yellow'      => '#dfb317',
  'yellowgreen' => '#a4a61d',
  'green'       => '#97ca00',
  'brightgreen' => '#4c1',
}

# http://shields.io/
$shield_template = <<EOT
<svg xmlns="http://www.w3.org/2000/svg" width="106" height="20">
  <linearGradient id="b" x2="0" y2="100%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <mask id="a">
    <rect width="106" height="20" rx="3" fill="#fff"/>
  </mask>
  <g mask="url(#a)">
    <path fill="#555" d="M0 0h63v20H0z"/>
    <path fill="__background__" d="M63 0h43v20H63z"/>
    <path fill="url(#b)" d="M0 0h106v20H0z"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
    <text x="31.5" y="15" fill="#010101" fill-opacity=".3">coverage</text>
    <text x="31.5" y="14">coverage</text>
    <text x="83.5" y="15" fill="#010101" fill-opacity=".3">__percent__%</text>
    <text x="83.5" y="14">__percent__%</text>
  </g>
</svg>
EOT

def get_shield_svg(code_coverage_percent)
  color = $colors['red']
  
  if code_coverage_percent >= 65.0 && code_coverage_percent < 79
    color = $colors['orange']
  elsif code_coverage_percent >= 80.0 && code_coverage_percent < 86
    color = $colors['yellow']
  elsif code_coverage_percent >= 87.0 && code_coverage_percent < 92
    color = $colors['yellowgreen']
  elsif code_coverage_percent >= 93.0 && code_coverage_percent < 100
    color = $colors['green']
  elsif code_coverage_percent == 100.0
    color = $colors['brightgreen']
  end
  
  template = $shield_template.gsub('__percent__', code_coverage_percent.ceil.to_s)
  template.gsub!('__background__', color)

  return template
end

total_reached = 0
total_statements = 0

json = JSON.load(File.read("coverage.json"))

json['Packages'].each do |package|
  package['Functions'].each do |function|
    next if function['Statements'].nil?

    function['Statements'].each do |statement|
      total_reached += 1 if statement['Reached'].to_i > 0
      total_statements += 1
    end
  end
end

percent = (total_reached.to_f / total_statements) * 100.0
shield  = get_shield_svg(percent)

File.open('coverage.svg', 'w') do |fh|
  fh.write(shield)
end

