# Métricas e KPIs

**Versão:** 1.0  
**Data:** 18 de Dezembro de 2025  
**Status:** Estabelecimento de Baseline

## Visão Geral

Este documento define métricas de sucesso, KPIs e dashboards para monitoramento contínuo do projeto. Métricas são categorizadas por desenvolvimento, qualidade, performance e negócio.

## Índice
1. [Métricas de Desenvolvimento](#métricas-de-desenvolvimento)
2. [Métricas de Qualidade](#métricas-de-qualidade)
3. [Métricas de Performance](#métricas-de-performance)
4. [Métricas de Negócio](#métricas-de-negócio)
5. [Dashboards](#dashboards)
6. [Alertas e SLAs](#alertas-e-slas)

---

## Métricas de Desenvolvimento

### Velocity

**Definição:** Story points completados por sprint  
**Objetivo:** Medir produtividade da equipe  
**Frequência:** Por sprint (2 semanas)

**Targets:**
| Time Size | Target Points/Sprint | Min Acceptable |
|-----------|---------------------|----------------|
| 2 devs | 25 | 20 |
| 3 devs | 35 | 28 |
| 4 devs | 45 | 36 |

**Tracking:**
```
Sprint X Velocity
━━━━━━━━━━━━━━━━━━━━━━
Planned:    45 points
Completed:  42 points  (93%)
Carry-over:  3 points

Trend (last 3 sprints):
Sprint 1: 38 points
Sprint 2: 40 points
Sprint 3: 42 points ↗️
```

**Alertas:**
- 🔴 Velocity < 80% do target
- 🟡 Velocity 80-90% do target
- 🟢 Velocity ≥ 90% do target

---

### Cycle Time

**Definição:** Tempo médio de uma história desde "In Progress" até "Done"  
**Objetivo:** Medir eficiência de entrega  
**Frequência:** Contínua

**Targets:**
| Story Size | Target Cycle Time | Max Acceptable |
|------------|------------------|----------------|
| 1-2 points | 1 dia | 2 dias |
| 3-5 points | 2-3 dias | 5 dias |
| 8 points | 4-5 dias | 7 dias |
| 13 points | 7-10 dias | 14 dias |

**Tracking:**
- Moving average (last 10 stories)
- P50, P75, P90 percentiles
- Trend over time

**Alertas:**
- 🔴 Cycle time > 150% do target
- 🟡 Cycle time 120-150% do target

---

### Lead Time

**Definição:** Tempo desde criação da issue até deploy em produção  
**Objetivo:** Medir time-to-market  
**Frequência:** Contínua

**Target:** < 2 semanas (P75)

**Breakdown:**
- Backlog → Ready: < 2 dias
- Ready → In Progress: < 1 dia
- In Progress → Review: < 5 dias
- Review → Done: < 2 dias
- Done → Deploy: < 3 dias

---

### Code Churn

**Definição:** % de código modificado após commit inicial  
**Objetivo:** Medir estabilidade e rework  
**Frequência:** Semanal

**Targets:**
- Churn < 20% (saudável)
- Churn 20-30% (atenção)
- Churn > 30% (problema)

**Tracking:**
```bash
git diff --shortstat HEAD~7..HEAD
# Calcular % de linhas modificadas vs. adicionadas
```

---

### Pull Request Metrics

**Definição:** Métricas de code review  
**Objetivo:** Garantir quality e collaboration  
**Frequência:** Contínua

**Targets:**
| Métrica | Target | Max |
|---------|--------|-----|
| Time to First Review | < 4 horas | 24h |
| Time to Merge | < 24 horas | 3 dias |
| PR Size | < 400 linhas | 800 linhas |
| Review Comments | 2-5 | 20 |

**Tracking:**
- Average time to review
- PR size distribution
- Review thoroughness (comments/LOC)

---

## Métricas de Qualidade

### Test Coverage

**Definição:** % de código coberto por testes  
**Objetivo:** Garantir qualidade e confiabilidade  
**Frequência:** Por commit (CI)

**Targets por Fase:**
| Fase | Target | Min Acceptable |
|------|--------|----------------|
| Fase 1 (Semanas 1-8) | 95% | 90% |
| Fase 2 (Semanas 9-16) | 98% | 95% |
| Fase 3 (Semanas 17-20) | 98% | 97% |

**Breakdown:**
- Unit tests: 80% do total
- Integration tests: 15% do total
- E2E tests: 5% do total

**Tracking:**
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

Total Coverage: 97.8%

By package:
internal/mcp/server:    99.2%
internal/elements:      98.5%
internal/portfolio:     96.3%
internal/security:      99.8%
```

**Alertas:**
- 🔴 Coverage drop > 2%
- 🔴 New code coverage < 90%
- 🟡 Coverage < target

---

### Bug Density

**Definição:** Bugs encontrados / 1000 LOC  
**Objetivo:** Medir qualidade do código  
**Frequência:** Por release

**Targets:**
- Fase 1: < 5 bugs/1000 LOC
- Fase 2: < 3 bugs/1000 LOC
- Fase 3: < 1 bug/1000 LOC

**Severity Breakdown:**
| Severity | Max Acceptable |
|----------|----------------|
| P0 (Critical) | 0 |
| P1 (High) | 2 |
| P2 (Medium) | 10 |
| P3 (Low) | ∞ |

---

### Defect Escape Rate

**Definição:** Bugs encontrados em produção / total de bugs  
**Objetivo:** Medir efetividade de testes  
**Frequência:** Por release

**Target:** < 5% (95% dos bugs encontrados pré-produção)

**Tracking:**
- Bugs found in dev: X
- Bugs found in testing: Y
- Bugs found in production: Z
- Escape rate: Z / (X + Y + Z)

---

### Technical Debt

**Definição:** Estimate de tempo para resolver dívida técnica  
**Objetivo:** Manter codebase saudável  
**Frequência:** Mensal

**Targets:**
- New tech debt: < 5% do velocity
- Total tech debt: < 20% de 1 sprint

**Tracking:**
- Issues marcadas com label "tech-debt"
- Story points estimados
- Age distribution

**Strategy:**
- Allocate 10-15% de cada sprint para tech debt
- Pay down before it compounds

---

### Code Quality Score

**Definição:** Score agregado de qualidade (linters, complexity, etc)  
**Objetivo:** Manter high code quality  
**Frequência:** Por PR

**Components:**
1. **golangci-lint:** 0 issues (weight: 40%)
2. **Cyclomatic complexity:** < 15 (weight: 30%)
3. **Code duplication:** < 5% (weight: 15%)
4. **Comment coverage:** > 50% public APIs (weight: 15%)

**Target:** Score ≥ 90/100

---

## Métricas de Performance

### Startup Time

**Definição:** Tempo de inicialização do servidor  
**Objetivo:** Fast startup para UX  
**Frequência:** Por build

**Targets:**
| Milestone | Target | Max |
|-----------|--------|-----|
| M0.1 | < 100ms | 200ms |
| M1 | < 75ms | 150ms |
| M2 | < 50ms | 100ms |
| M3 | < 50ms | 75ms |

**Tracking:**
```bash
time ./bin/mcp-server --version
# real    0m0.042s
```

---

### Memory Footprint

**Definição:** Uso de memória em diferentes cargas  
**Objetivo:** Efficient resource usage  
**Frequência:** Por release

**Targets:**
| Scenario | Target | Max |
|----------|--------|-----|
| Idle (no elements) | < 30MB | 50MB |
| 100 elements loaded | < 50MB | 75MB |
| 1,000 elements | < 100MB | 150MB |
| 10,000 elements | < 200MB | 300MB |

**Tracking:**
```bash
# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# Runtime memory
ps aux | grep mcp-server
# RSS (Resident Set Size)
```

---

### Operation Latency

**Definição:** Tempo de resposta para operações MCP  
**Objetivo:** Fast, responsive server  
**Frequência:** Contínua (benchmarks)

**Targets (P95):**
| Operation | Target | Max |
|-----------|--------|-----|
| list_elements (100) | < 5ms | 10ms |
| list_elements (1000) | < 10ms | 20ms |
| get_element | < 1ms | 3ms |
| create_element | < 5ms | 10ms |
| search_elements (1000) | < 10ms | 25ms |
| github_sync (100) | < 5s | 10s |

**Tracking:**
```go
// Benchmark tests
func BenchmarkListElements(b *testing.B) {
    for i := 0; i < b.N; i++ {
        handler.ListElements(ctx, input)
    }
}
```

**Alertas:**
- 🔴 P95 > Max
- 🟡 P95 > Target

---

### Throughput

**Definição:** Operações processadas por segundo  
**Objetivo:** High concurrency capability  
**Frequência:** Load testing

**Targets:**
- Single operation: > 1000 req/s
- Mixed operations: > 500 req/s
- Under load (10k elements): > 200 req/s

**Tracking:**
- Load testing tools (Apache Bench, vegeta)
- Concurrent goroutines
- Resource saturation point

---

### Resource Utilization

**Definição:** CPU, Memory, I/O usage under load  
**Objetivo:** Efficient resource usage  
**Frequência:** Load testing

**Targets:**
- CPU: < 50% single core (normal load)
- Memory: < 200MB (10k elements)
- I/O wait: < 10%
- Goroutines: < 1000 (normal load)

**Tracking:**
```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Runtime metrics
runtime.NumGoroutine()
runtime.ReadMemStats(&m)
```

---

## Métricas de Negócio

### Feature Adoption

**Definição:** % de usuários usando cada feature  
**Objetivo:** Validar product-market fit  
**Frequência:** Mensal (pós-release)

**Tracking (se telemetry opt-in):**
- Tool usage frequency
- Most/least used elements
- User workflows

**Target:** Core features > 80% adoption

---

### Time to Value

**Definição:** Tempo até usuário extrair valor  
**Objetivo:** Great onboarding experience  
**Frequência:** User research

**Targets:**
- First element created: < 5 minutos
- First GitHub sync: < 15 minutos
- Active usage (10+ elements): < 1 dia

---

### User Satisfaction

**Definição:** Feedback qualitativo e quantitativo  
**Objetivo:** Happy users  
**Frequência:** Quarterly surveys

**Metrics:**
- NPS (Net Promoter Score): > 50
- CSAT (Customer Satisfaction): > 4/5
- GitHub stars: > 100 (6 meses)
- Issues closed / opened: > 2

---

### Community Engagement

**Definição:** Contribuições e atividade da comunidade  
**Objetivo:** Healthy open-source project  
**Frequência:** Mensal

**Metrics:**
- Contributors: > 10 (1 ano)
- Pull requests: > 5/month
- Issues reported: > 10/month
- Discord/Slack members: > 100

---

## Dashboards

### Development Dashboard

**URL:** `https://github.com/fsvxavier/nexs-mcp/actions`  
**Update:** Real-time

**Widgets:**
1. **Build Status**
   - CI pipeline status (pass/fail)
   - Last successful build
   - Build duration trend

2. **Test Coverage**
   - Current coverage %
   - Coverage trend (last 30 days)
   - Coverage by package

3. **Velocity**
   - Current sprint progress
   - Burndown chart
   - Velocity trend

4. **Code Quality**
   - Linting issues
   - Code smells
   - Technical debt

---

### Quality Dashboard

**URL:** `codecov.io/gh/fsvxavier/nexs-mcp`  
**Update:** Per commit

**Widgets:**
1. **Test Coverage**
   - Overall coverage %
   - Coverage diff (vs. main)
   - Uncovered lines

2. **Bug Metrics**
   - Open bugs by severity
   - Bug trend (opened vs. closed)
   - Mean time to resolution

3. **Security**
   - Vulnerability count
   - Security score
   - Dependency alerts

---

### Performance Dashboard

**Tools:** pprof, Grafana (optional)  
**Update:** Per benchmark run

**Widgets:**
1. **Latency**
   - P50, P95, P99 latencies
   - Latency distribution
   - Trend over releases

2. **Resource Usage**
   - Memory usage
   - CPU usage
   - Goroutine count

3. **Throughput**
   - Requests per second
   - Concurrent operations
   - Resource saturation

---

### Release Dashboard

**URL:** GitHub Releases  
**Update:** Per release

**Metrics:**
1. **Release Health**
   - Download count
   - Installation success rate
   - Crash reports

2. **Adoption**
   - Active users
   - Feature usage
   - Upgrade rate

3. **Feedback**
   - GitHub stars
   - Issue volume
   - PR contributions

---

## Alertas e SLAs

### Alert Levels

**🔴 Critical (P0):**
- Action: Immediate attention
- Response: < 1 hora
- Examples:
  - Build broken
  - Security vulnerability (CVSS > 8)
  - Production crash
  - Data loss

**🟠 High (P1):**
- Action: Same day
- Response: < 4 horas
- Examples:
  - Test coverage drop > 2%
  - Performance regression > 20%
  - Multiple test failures

**🟡 Medium (P2):**
- Action: Within 3 dias
- Response: < 1 dia
- Examples:
  - Linting issues
  - Minor performance regression
  - Non-critical bug

**🟢 Low (P3):**
- Action: Next sprint
- Response: Best effort
- Examples:
  - Tech debt
  - Nice-to-have improvements

---

### SLA Definitions

#### Development SLAs

| Metric | SLA | Breach Action |
|--------|-----|---------------|
| Build time | < 5 min | Optimize CI |
| Test execution | < 60s | Parallelize tests |
| PR review time | < 24h | Add reviewers |
| Bug fix (P0) | < 24h | Escalate |
| Bug fix (P1) | < 1 week | Re-prioritize |

#### Quality SLAs

| Metric | SLA | Breach Action |
|--------|-----|---------------|
| Test coverage | > 95% | Block merge |
| Security vulnerabilities | 0 critical | Immediate fix |
| Linting issues | 0 | Block merge |
| Code review approval | 2+ reviewers | Add reviewer |

#### Performance SLAs

| Metric | SLA | Breach Action |
|--------|-----|---------------|
| Startup time | < 50ms | Profiling session |
| Memory footprint | < 50MB idle | Memory audit |
| Operation latency | < P95 target | Optimization sprint |

---

### Alert Configuration

```yaml
# Example alert configuration
alerts:
  build_failure:
    severity: critical
    notify: ["tech-lead", "slack-channel"]
    sla: 1h
    
  coverage_drop:
    severity: high
    threshold: 2%
    notify: ["tech-lead"]
    sla: 4h
    
  performance_regression:
    severity: high
    threshold: 20%
    notify: ["tech-lead", "senior-devs"]
    sla: 24h
```

---

## Tracking Tools

### Recommended Tools

1. **GitHub Projects**
   - Kanban board
   - Sprint planning
   - Issue tracking

2. **GitHub Actions**
   - CI/CD metrics
   - Build times
   - Test results

3. **Codecov**
   - Test coverage
   - Coverage trends
   - PR coverage diff

4. **golangci-lint**
   - Code quality
   - Linting issues
   - Complexity metrics

5. **pprof** (Go native)
   - CPU profiling
   - Memory profiling
   - Goroutine analysis

6. **Grafana** (optional)
   - Custom dashboards
   - Real-time metrics
   - Alerting

---

## Reporting Cadence

### Daily
- Standup metrics:
  - Blockers
  - Yesterday's completions
  - Today's plan

### Weekly
- Sprint health:
  - Velocity
  - Burndown
  - Blockers
  - Risk status

### Bi-weekly (End of Sprint)
- Sprint retrospective:
  - Velocity achieved
  - What went well
  - What to improve
  - Action items

### Monthly
- Project health:
  - Milestone progress
  - Quality metrics
  - Performance benchmarks
  - Risk review

### Per Release
- Release report:
  - Features delivered
  - Bugs fixed
  - Performance comparison
  - Test coverage
  - Known issues

---

## Success Criteria

### M1: Foundation Complete (Semana 8)

| Categoria | Métrica | Target | Status |
|-----------|---------|--------|--------|
| Development | Velocity | 180 points (8 weeks) | TBD |
| Development | Features | 3 element types | TBD |
| Quality | Test coverage | > 95% | TBD |
| Quality | Bugs (P0/P1) | 0 open | TBD |
| Performance | Startup time | < 75ms | TBD |
| Performance | Memory | < 50MB | TBD |

### M2: Feature Complete (Semana 16)

| Categoria | Métrica | Target | Status |
|-----------|---------|--------|--------|
| Development | Features | 6 element types + 49 tools | TBD |
| Quality | Test coverage | > 98% | TBD |
| Quality | Security score | 100/100 | TBD |
| Performance | All targets | 100% met | TBD |

### M3: Production Ready (Semana 20)

| Categoria | Métrica | Target | Status |
|-----------|---------|--------|--------|
| Quality | Zero P0/P1 bugs | ✅ | TBD |
| Performance | Benchmarks | All pass | TBD |
| Documentation | Coverage | 100% APIs | TBD |
| Release | Artifacts | All platforms | TBD |

---

**Última Atualização:** 18 de Dezembro de 2025  
**Próxima Revisão:** Semanal  
**Owner:** Tech Lead + Project Manager
