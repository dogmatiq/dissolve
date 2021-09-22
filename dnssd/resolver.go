package dnssd

// // unicastResolver is an implementation of Resolver that uses conventional
// // unicast DNS queries.
// type unicastResolver struct {
// 	Client *dns.Client
// 	Config *dns.ClientConfig
// }

// // newUnicastResolver returns a new unicast resolver that uses the DNS servers
// // defined in /etc/resolv.conf
// func newUnicastResolver() (*unicastResolver, error) {
// 	conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &unicastResolver{
// 		&dns.Client{},
// 		conf,
// 	}, nil
// }

// // Browse for all services of a given type in a given domain.
// func (r *unicastResolver) Browse(ctx context.Context, service, domain string, entries chan<- *zeroconf.ServiceEntry) error {
// 	defer close(entries)

// 	name := service + "." + domain
// 	msg, err := r.query(ctx, name, dns.TypePTR)
// 	if err != nil {
// 		return err
// 	} else if msg == nil {
// 		// if no DNS message is returned from any of the DNS servers there can
// 		// be no services, which is not an error.
// 		return nil
// 	}

// 	wg, subCtx := errgroup.WithContext(ctx)

// 	for _, a := range msg.Answer {
// 		if ptr, ok := a.(*dns.PTR); ok {
// 			instance := strings.TrimSuffix(
// 				ptr.Ptr, "."+ptr.Hdr.Name,
// 			)

// 			sr := zeroconf.NewServiceRecord(instance, service, domain)

// 			wg.Go(func() error {
// 				return r.queryInstance(subCtx, sr, entries)
// 			})
// 		}
// 	}

// 	return wg.Wait()
// }

// // Lookup a specific service by its name and type in a given domain.
// func (r *unicastResolver) Lookup(ctx context.Context, instance, service, domain string, entries chan<- *zeroconf.ServiceEntry) error {
// 	defer close(entries)

// 	sr := zeroconf.NewServiceRecord(instance, service, domain)
// 	return r.queryInstance(ctx, sr, entries)
// }

// func (r *unicastResolver) queryInstance(
// 	ctx context.Context,
// 	sr *zeroconf.ServiceRecord,
// 	entries chan<- *zeroconf.ServiceEntry,
// ) error {
// 	entry := &zeroconf.ServiceEntry{
// 		ServiceRecord: *sr,
// 	}

// 	wg, subCtx := errgroup.WithContext(ctx)

// 	for _, t := range []uint16{dns.TypeSRV, dns.TypeTXT} {
// 		qType := t
// 		wg.Go(func() error {
// 			msg, err := r.query(subCtx, sr.ServiceInstanceName(), qType)
// 			if err != nil {
// 				return err
// 			}
// 			unpackServiceEntry(entry, msg)
// 			return nil
// 		})
// 	}

// 	if err := wg.Wait(); err != nil {
// 		return err
// 	}

// 	if entry.HostName != "" {
// 		select {
// 		case entries <- entry:
// 		case <-ctx.Done():
// 			return ctx.Err()
// 		}
// 	}

// 	return nil
// }

// func (r *unicastResolver) query(ctx context.Context, n string, t uint16) (*dns.Msg, error) {
// 	req := &dns.Msg{}
// 	req.SetQuestion(n, t)

// 	for _, s := range r.Config.Servers {
// 		addr := net.JoinHostPort(s, r.Config.Port)
// 		res, _, err := r.Client.ExchangeContext(ctx, req, addr)

// 		if err != nil {
// 			return nil, err
// 		} else if res == nil {
// 			continue // server did not have a response for this query
// 		} else if res.Rcode == dns.RcodeNameError || res.Rcode == dns.RcodeSuccess {
// 			return res, nil
// 		}
// 	}

// 	return nil, nil
// }

// func unpackServiceEntry(entry *zeroconf.ServiceEntry, msg *dns.Msg) {
// 	for _, a := range msg.Answer {
// 		switch rec := a.(type) {
// 		case *dns.SRV:
// 			entry.HostName = rec.Target
// 			entry.Port = int(rec.Port)
// 			entry.TTL = rec.Hdr.Ttl
// 		case *dns.TXT:
// 			entry.Text = append(entry.Text, rec.Txt...)
// 		}
// 	}
// }
